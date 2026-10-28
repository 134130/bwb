package vault

import (
	"context"
	"errors"
	"time"

	"github.com/134130/bwb/internal/bitwarden"
	"github.com/google/uuid"
)

type RepromptRequired struct {
	Profile string `json:"profile"`
	Item    string `json:"item"`
}

func (e *RepromptRequired) Unwrap() error { return bitwarden.ErrReprompt }
func (e *RepromptRequired) Error() string { return "item requires master password reprompt" }

type RepromptRequest struct {
	Profile  string `json:"profile"`
	Item     string `json:"item"`
	Password []byte `json:"password"`
}
type RepromptResponse struct {
	Token string `json:"token"`
}
type RepromptWatchRequest struct {
	Profile string `json:"profile"`
	Token   string `json:"token"`
}
type repromptGrant struct {
	item    string
	expires time.Time
	watch   context.Context
	done    chan struct{}
}

// repromptMu also excludes lock cleanup while config reload invalidates grants.
func (p *profile) clearReprompt() {
	p.repromptMu.Lock()
	defer p.repromptMu.Unlock()
	for _, g := range p.reprompts {
		close(g.done)
	}
	p.reprompts = nil
}

func (s *Service) Reprompt(ctx context.Context, req RepromptRequest) (RepromptResponse, error) {
	defer clear(req.Password)
	ctx, p, done, err := s.begin(ctx, req.Profile, true)
	if err != nil {
		return RepromptResponse{}, err
	}
	defer done()
	if p.session == nil {
		return RepromptResponse{}, s.authRequired(ctx, p, req.Profile, "unlock")
	}
	required, err := p.snapshot.RequiresReprompt(req.Item)
	if err != nil {
		return RepromptResponse{}, err
	}
	if !required {
		return RepromptResponse{}, errors.New("item does not require reprompt")
	}
	// A password never overrides organization restrictions.
	allowed, err := p.snapshot.AllowReprompt(req.Item)
	if err != nil {
		return RepromptResponse{}, err
	}
	if _, err = p.session.Describe(ctx, allowed, req.Item); err != nil {
		return RepromptResponse{}, err
	}
	if err = p.session.VerifyPassword(ctx, req.Password); err != nil {
		return RepromptResponse{}, err
	}
	if err = s.touch(ctx, p); err != nil {
		return RepromptResponse{}, err
	}
	p.repromptMu.Lock()
	defer p.repromptMu.Unlock()
	for token, g := range p.reprompts {
		if g.watch != nil && g.watch.Err() != nil || g.watch == nil && !time.Now().Before(g.expires) {
			close(g.done)
			delete(p.reprompts, token)
		}
	}
	if len(p.reprompts) >= 32 {
		return RepromptResponse{}, errors.New("too many active reprompt grants")
	}
	if p.reprompts == nil {
		p.reprompts = map[string]*repromptGrant{}
	}
	token := uuid.NewString()
	p.reprompts[token] = &repromptGrant{item: req.Item, expires: time.Now().Add(AuthTimeout), done: make(chan struct{})}
	return RepromptResponse{Token: token}, nil
}

// WatchReprompt keeps this capability alive only while its owner remains connected.
func (s *Service) WatchReprompt(ctx context.Context, req RepromptWatchRequest) (Empty, error) {
	_, p, done, err := s.begin(ctx, req.Profile, false)
	if err != nil {
		return Empty{}, err
	}
	p.repromptMu.Lock()
	g := p.reprompts[req.Token]
	if g == nil || g.watch != nil || !time.Now().Before(g.expires) {
		p.repromptMu.Unlock()
		done()
		return Empty{}, ErrStateChanged
	}
	g.watch = ctx
	p.repromptMu.Unlock()
	done()
	select {
	case <-ctx.Done():
	case <-g.done:
	}
	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
	defer cancel()
	_, p, done, err = s.begin(cleanup, req.Profile, false)
	if err == nil {
		p.repromptMu.Lock()
		if p.reprompts[req.Token] == g {
			close(g.done)
			delete(p.reprompts, req.Token)
		}
		p.repromptMu.Unlock()
		done()
	}
	return Empty{}, ErrStateChanged
}

func (s *Service) itemSnapshot(ctx context.Context, p *profile, profile, id, token string) (*bitwarden.Snapshot, error) {
	required, err := p.snapshot.RequiresReprompt(id)
	if err != nil {
		return nil, err
	}
	if !required {
		return p.snapshot, nil
	}
	allowed := s.policies[profile].InsecureSkipReprompt
	p.repromptMu.Lock()
	if g := p.reprompts[token]; g != nil && g.item == id {
		allowed = allowed || g.watch != nil && g.watch.Err() == nil || g.watch == nil && time.Now().Before(g.expires)
	}
	p.repromptMu.Unlock()
	if !allowed {
		return nil, &RepromptRequired{Profile: profile, Item: id}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return p.snapshot.AllowReprompt(id)
}
