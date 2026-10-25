package cli

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/134130/bwb/internal/vault"
	"github.com/134130/bwb/internal/vaultapi"
	"github.com/spf13/cobra"
)

func listCommand(kind string, profile *string, nonInteractive *bool) *cobra.Command {
	parent := &cobra.Command{Use: kind, Short: "Browse cached " + kind + " metadata"}
	req := vaultapi.ListRequest{Kind: kind}
	var format, group string
	list := &cobra.Command{Use: "list", Short: "List cached metadata without contacting the server", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if format != "table" && format != "json" {
				return errors.New("format must be table or json")
			}
			if group != "" && group != "folder" && group != "organization" && group != "collection" {
				return errors.New("group-by must be folder, organization, or collection")
			}
			req.Profile = *profile
			if err := vault.ValidateList(req); err != nil {
				return err
			}
			client, err := connectCommand(cmd)
			if err != nil {
				return err
			}
			defer client.Close()
			execution := executor{client: client, prompt: interactivePrompt(cmd, *nonInteractive)}
			entries, err := execution.List(cmd.Context(), req)
			if err != nil {
				return err
			}
			if err := cmd.Context().Err(); err != nil {
				return err
			}
			return writeList(cmd.OutOrStdout(), kind, format, group, entries)
		}}
	list.Flags().StringVar(&req.Search, "search", "", "Case-insensitive substring in names")
	list.Flags().StringSliceVar(&req.Sort, "sort", []string{"name"}, "Sort keys, comma-separated (name; items also type,folder,organization)")
	list.Flags().BoolVar(&req.Reverse, "reverse", false, "Reverse sort keys; ties retain ID order")
	list.Flags().StringVar(&format, "format", "table", "Output format: table or json")
	if kind == "item" || kind == "collection" {
		list.Flags().StringVar(&req.OrganizationID, "organization", "", "Filter by organization UUID")
	}
	if kind == "item" {
		list.Flags().StringVar(&req.FolderID, "folder", "", "Filter by folder UUID")
		list.Flags().StringVar(&req.CollectionID, "collection", "", "Filter by collection UUID")
		list.Flags().StringVar(&group, "group-by", "", "Group table by folder, organization, or collection (JSON stays flat)")
	}
	parent.AddCommand(list)
	return parent
}

// Escape control/format characters, including terminal escapes, without clipping names or IDs.
func display(value string) string {
	quoted := strconv.QuoteToGraphic(value)
	return quoted[1 : len(quoted)-1]
}

func writeList(out io.Writer, kind, format, group string, entries []vaultapi.ListEntry) error {
	if format == "json" {
		if entries == nil {
			entries = []vaultapi.ListEntry{}
		}
		return json.NewEncoder(out).Encode(struct {
			Version int                  `json:"version"`
			Data    []vaultapi.ListEntry `json:"data"`
		}{1, entries})
	}
	if group == "" || len(entries) == 0 {
		return writeListTable(out, kind, entries)
	}
	type section struct {
		id, name string
		entries  []vaultapi.ListEntry
	}
	sections := map[string]*section{}
	for _, entry := range entries {
		refs := []vaultapi.NamedID{}
		switch group {
		case "folder":
			if entry.Folder != nil {
				refs = append(refs, *entry.Folder)
			}
		case "organization":
			if entry.Organization != nil {
				refs = append(refs, *entry.Organization)
			}
		case "collection":
			refs = entry.Collections
		}
		if len(refs) == 0 {
			refs = []vaultapi.NamedID{{Name: "—"}}
		}
		for _, ref := range refs {
			if sections[ref.ID] == nil {
				sections[ref.ID] = &section{id: ref.ID, name: ref.Name}
			}
			sections[ref.ID].entries = append(sections[ref.ID].entries, entry)
		}
	}
	ordered := make([]*section, 0, len(sections))
	for _, section := range sections {
		ordered = append(ordered, section)
	}
	slices.SortFunc(ordered, func(a, b *section) int {
		if order := cmp.Compare(strings.ToLower(a.name), strings.ToLower(b.name)); order != 0 {
			return order
		}
		return cmp.Compare(a.id, b.id)
	})
	for i, section := range ordered {
		if i != 0 {
			if _, err := fmt.Fprintln(out); err != nil {
				return err
			}
		}
		label := display(section.name)
		if section.id != "" {
			label += " [" + section.id + "]"
		}
		if _, err := fmt.Fprintf(out, "%s: %s\n", group, label); err != nil {
			return err
		}
		if err := writeListTable(out, kind, section.entries); err != nil {
			return err
		}
	}
	return nil
}

func writeListTable(out io.Writer, kind string, entries []vaultapi.ListEntry) error {
	table := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	header := "ID\tNAME"
	switch kind {
	case "item":
		header = "ID\tTYPE\tORGANIZATION\tFOLDER\tCOLLECTIONS\tFAV\tNAME"
	case "collection":
		header = "ID\tORGANIZATION\tNAME"
	}
	if _, err := fmt.Fprintln(table, header); err != nil {
		return err
	}
	name := func(ref *vaultapi.NamedID) string {
		if ref == nil {
			return "—"
		}
		return display(ref.Name)
	}
	for _, entry := range entries {
		cells := []string{entry.ID}
		switch kind {
		case "item":
			names := make([]string, 0, len(entry.Collections))
			for _, collection := range entry.Collections {
				names = append(names, display(collection.Name))
			}
			slices.Sort(names)
			collections := strings.Join(names, ", ")
			if collections == "" {
				collections = "—"
			}
			favorite := "—"
			if entry.Favorite {
				favorite = "*"
			}
			cells = append(cells, display(entry.Type), name(entry.Organization), name(entry.Folder), collections, favorite)
		case "collection":
			cells = append(cells, name(entry.Organization))
		}
		cells = append(cells, display(entry.Name))
		if _, err := fmt.Fprintln(table, strings.Join(cells, "\t")); err != nil {
			return err
		}
	}
	return table.Flush()
}
