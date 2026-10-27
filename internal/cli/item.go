package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/134130/bwb/internal/tui"
	"golang.org/x/term"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/134130/bwb/internal/vaultapi"
	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

func itemCommand(totp bool, profile *string, nonInteractive *bool) *cobra.Command {
	req := vaultapi.ItemRequest{TOTP: totp}
	var format string
	var syncFirst, copyField bool
	use, short := "read ITEM", "Show an item and its field keys; select one value with --field"
	if totp {
		use, short = "totp ITEM", "Generate the current TOTP code without returning its seed"
	}
	cmd := &cobra.Command{Use: use, Short: short, Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		if copyField && (req.Field == "" || req.Reveal || format != "text") {
			return errors.New("--copy requires --field and cannot be combined with --reveal or --format json")
		}
		if format != "text" && format != "json" {
			return errors.New("format must be text or json")
		}
		if cmd.Flags().Changed("field") && req.Field == "" {
			return errors.New("--field requires a field key; omit it to see the item")
		}
		if req.Reveal && req.Field != "" {
			return errors.New("--field already returns its value; omit --reveal")
		}
		client, err := connectCommand(cmd)
		if err != nil {
			return err
		}
		defer client.Close()
		execution := executor{Client: client, Prompt: interactivePrompt(cmd, *nonInteractive)}
		if syncFirst {
			if _, err := execution.Sync(cmd.Context(), vaultapi.ProfileRequest{Profile: *profile}); err != nil {
				return err
			}
		}
		if totp && format == "text" && !*nonInteractive {
			input, inOK := cmd.InOrStdin().(*os.File)
			output, outOK := cmd.OutOrStdout().(*os.File)
			if inOK && outOK && term.IsTerminal(int(input.Fd())) && term.IsTerminal(int(output.Fd())) {
				return tui.RunTOTP(cmd.Context(), *profile, args[0], client, connectContext, input, output)
			}
		}
		req.Profile, req.Item = *profile, args[0]
		result, err := execution.Item(cmd.Context(), req)
		if err != nil {
			return err
		}
		if copyField {
			return copyItem(cmd.Context(), cmd.ErrOrStderr(), result, clipboard.WriteAll)
		}
		return writeItem(cmd.OutOrStdout(), format, *profile, result, req.Field != "" || totp)
	}}
	cmd.Long = short + ".\nITEM is a UUID or an exact case-insensitive name. Duplicate names require an ID.\nReads use the cached vault; pass --sync to refresh first."
	cmd.Flags().StringVar(&format, "format", "text", "Output format: text or json")
	cmd.Flags().BoolVar(&syncFirst, "sync", false, "Sync before reading")
	if !totp {
		cmd.Flags().BoolVar(&copyField, "copy", false, "Copy the selected field instead of printing it")
		cmd.Flags().StringVar(&req.Field, "field", "", "Return one field value; omit to discover available keys")
		cmd.Flags().BoolVar(&req.Reveal, "reveal", false, "Show hidden values in the item details")
		cmd.Example = "  bwb item read 'Example account'\n  bwb item read ITEM_ID --field password\n  bwb item read ITEM_ID --reveal --format json"
	}
	return cmd
}

func itemType(kind string) string {
	labels := map[string]string{"login": "Login credentials", "secureNote": "Secure note", "card": "Card", "identity": "Identity", "sshKey": "SSH key", "bankAccount": "Bank account", "driversLicense": "Driver's license", "passport": "Passport"}
	if label := labels[kind]; label != "" {
		return label
	}
	return kind
}

func shellQuote(value string) string { return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'" }

func writeItem(out io.Writer, format, profile string, result vaultapi.ItemResponse, single bool) error {
	if format == "json" {
		return json.NewEncoder(out).Encode(struct {
			Version int                   `json:"version"`
			Data    vaultapi.ItemResponse `json:"data"`
		}{1, result})
	}
	if single {
		if len(result.Fields) != 1 || result.Fields[0].Value == nil {
			return errors.New("item response has no field value")
		}
		_, err := io.WriteString(out, *result.Fields[0].Value)
		return err
	}
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "%s\nType: %s\nID: %s\n", display(result.Item.Name), display(itemType(result.Item.Type)), result.Item.ID)
	if result.Item.Organization != nil {
		fmt.Fprintf(&buffer, "Organization: %s\n", display(result.Item.Organization.Name))
	}
	if result.Item.Folder != nil {
		fmt.Fprintf(&buffer, "Folder: %s\n", display(result.Item.Folder.Name))
	}
	for _, collection := range result.Item.Collections {
		fmt.Fprintf(&buffer, "Collection: %s\n", display(collection.Name))
	}
	if result.Restriction != "" {
		fmt.Fprintf(&buffer, "\nFields unavailable: %s\n", display(result.Restriction))
	} else {
		buffer.WriteByte('\n')
		table := tabwriter.NewWriter(&buffer, 0, 4, 2, ' ', 0)
		fmt.Fprintln(table, "FIELD\tLABEL\tVALUE")
		var hidden, hasTOTP bool
		example := "name"
		for _, field := range result.Fields {
			value := "[hidden]"
			if field.State == "empty" {
				value = "[empty]"
			} else if field.Value != nil {
				value = display(*field.Value)
			}
			fmt.Fprintf(table, "%s\t%s\t%s\n", display(field.Key), display(field.Label), value)
			hidden = hidden || field.State == "hidden"
			hasTOTP = hasTOTP || field.Key == "totp-secret"
			if example == "name" {
				example = field.Key
			}
		}
		_ = table.Flush()
		prefix := "bwb --profile " + shellQuote(profile) + " item "
		fmt.Fprintf(&buffer, "\nRead a value: %sread %s --field %s\n", prefix, result.Item.ID, shellQuote(example))
		if hidden {
			fmt.Fprintf(&buffer, "Show hidden values: %sread %s --reveal\n", prefix, result.Item.ID)
		}
		if hasTOTP {
			fmt.Fprintf(&buffer, "Authentication code: %stotp %s\n", prefix, result.Item.ID)
		}
	}
	_, err := out.Write(buffer.Bytes())
	return err
}

func copyItem(ctx context.Context, out io.Writer, result vaultapi.ItemResponse, write func(string) error) error {
	if len(result.Fields) != 1 || result.Fields[0].Value == nil {
		return errors.New("item response has no field value")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := write(*result.Fields[0].Value); err != nil {
		return errors.New("clipboard unavailable")
	}
	_, err := fmt.Fprintln(out, "Copied")
	return err
}
