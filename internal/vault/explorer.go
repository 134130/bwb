package vault

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"

	"github.com/134130/bwb/internal/bitwarden"
)

type DescribeRequest struct {
	Profile string `json:"profile"`
	ID      string `json:"id"`
}
type DescribeResponse struct {
	Fields []bitwarden.FieldInfo `json:"fields"`
}
type ProfilesResponse struct {
	Names []string `json:"names"`
}

func (s *Service) Describe(ctx context.Context, req DescribeRequest) (DescribeResponse, error) {
	ctx, p, done, err := s.begin(ctx, req.Profile, true)
	if err != nil {
		return DescribeResponse{}, err
	}
	defer done()
	if p.session == nil {
		return DescribeResponse{}, s.authRequired(ctx, p, req.Profile, "unlock")
	}
	fields, err := p.session.Describe(ctx, p.snapshot, req.ID)
	if err != nil {
		return DescribeResponse{}, err
	}
	encoded, err := json.Marshal(fields)
	if err != nil {
		return DescribeResponse{}, err
	}
	if len(encoded) > 48000 {
		return DescribeResponse{}, errors.New("field metadata exceeds RPC limit")
	}
	if err := ctx.Err(); err != nil {
		return DescribeResponse{}, err
	}
	return DescribeResponse{Fields: fields}, nil
}

func (s *Service) Profiles(ctx context.Context, _ Empty) (ProfilesResponse, error) {
	keys, err := s.Store.List(ctx, "")
	if err != nil {
		return ProfilesResponse{}, err
	}
	names := map[string]bool{}
	for _, key := range keys {
		if name, ok := strings.CutSuffix(key, "/count"); ok && validProfile(name) == nil {
			names[name] = true
		}
	}
	s.configMu.RLock()
	for name := range s.policies {
		names[name] = true
	}
	s.configMu.RUnlock()
	result := ProfilesResponse{Names: []string{}}
	for name := range names {
		result.Names = append(result.Names, name)
	}
	slices.Sort(result.Names)
	return result, nil
}

// Activity is explicit user interaction. Background discovery and polling never call it.
func (s *Service) Activity(ctx context.Context, req ProfileRequest) (Empty, error) {
	ctx, p, done, err := s.begin(ctx, req.Profile, true)
	if err != nil {
		return Empty{}, err
	}
	defer done()
	if p.session == nil {
		return Empty{}, s.authRequired(ctx, p, req.Profile, "unlock")
	}
	return Empty{}, s.touch(ctx, p)
}
