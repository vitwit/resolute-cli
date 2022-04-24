package cmd

import (
	"fmt"
	"strings"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/spf13/cobra"
	tmtypes "github.com/tendermint/tendermint/types"
	"github.com/vitwit/resolute/client/query"
)

const (
	flagEvents = "events"
	flagType   = "type"

	typeHash   = "hash"
	typeAccSeq = "acc_seq"
	typeSig    = "signature"

	eventFormat = "{eventType}.{eventAttribute}={value}"
)

// authQueryTxCmd command to query the transaction by hash
func authQeuryTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx [hash]",
		Short: "Query for a transaction by hash",
		Long: strings.TrimSpace(fmt.Sprintf(`
	Example:
	$ resolute query tx <hash>`),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			cl.SetConfig()
			if args[0] == "" {
				return fmt.Errorf("argument should be a tx hash")
			}
			res, err := cl.QueryTx(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			if res.Empty() {
				return fmt.Errorf("no transaction found with hash %s", args[0])
			}
			return cl.PrintObject(res)

		},
	}
	return cmd
}

// QueryTxsByEventsCmd to query for paginated transactions that match a set of events
func QueryTxsByEventsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "txs",
		Short: "Query for paginated transactions that match a set of events",
		Long: strings.TrimSpace(
			fmt.Sprintf(`
Search for transactions that match the exact given events where results are paginated.
Each event takes the form of '{eventType}.{eventAttribute}={value}'. Please refer
to each module's documentation for the full set of events to query for. Each module
documents its respective events under 'xx_events.md'.

Example:
$ resolute query txs --events 'message.sender=cosmos1...&message.action=withdraw_delegator_reward' --page 1 --limit 30`),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			eventsRaw, _ := cmd.Flags().GetString(flagEvents)
			eventsStr := strings.Trim(eventsRaw, "'")

			var events []string
			if strings.Contains(eventsStr, "&") {
				events = strings.Split(eventsStr, "&")
			} else {
				events = append(events, eventsStr)
			}

			var tmEvents []string

			for _, event := range events {
				if !strings.Contains(event, "=") {
					return fmt.Errorf("invalid event; event %s should be of the format: %s", event, eventFormat)
				} else if strings.Count(event, "=") > 1 {
					return fmt.Errorf("invalid event; event %s should be of the format: %s", event, eventFormat)
				}

				tokens := strings.Split(event, "=")
				if tokens[0] == tmtypes.TxHeightKey {
					event = fmt.Sprintf("%s=%s", tokens[0], tokens[1])
				} else {
					event = fmt.Sprintf("%s='%s'", tokens[0], tokens[1])
				}

				tmEvents = append(tmEvents, event)

			}
			page, _ := cmd.Flags().GetInt(flags.FlagPage)
			limit, _ := cmd.Flags().GetInt(flags.FlagLimit)

			txs, err := cl.QueryTxsByEvents(cmd.Context(), tmEvents, page, limit, "")
			if err != nil {
				return err
			}

			return cl.PrintObject(txs)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().Int(flags.FlagPage, rest.DefaultPage, "Query a specific page of paginated results")
	cmd.Flags().Int(flags.FlagLimit, rest.DefaultLimit, "Query number of transactions results per page returned")
	cmd.Flags().String(flagEvents, "", fmt.Sprintf("list of transaction events in the form of %s", eventFormat))
	cmd.MarkFlagRequired(flagEvents)
	return cmd
}

// QUeryAccountCmd to query an account by address
func QUeryAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account [address]",
		Short: "QUery for account by address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			key, err := cl.DecodeBech32AccAddr(args[0])
			if err != nil {
				return err
			}
			cq := query.Query{Client: cl, Options: query.DefaultOptions()}

			res, err := cq.QueryAccount(key.String())
			if err != nil {
				return err
			}
			return cl.PrintObject(res)
		},
	}
	return cmd
}

// QueryAccountsCmd to query all accounts
func QueryAccountsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accounts",
		Short: "Query all the accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()

			pr, err := sdkclient.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			height, err := ReadHeight(cmd.Flags())
			if err != nil {
				return err
			}
			options := query.QueryOptions{Pagination: pr, Height: height}
			query := query.Query{cl, &options}

			res, err := query.QueryAccounts()
			if err != nil {
				return err
			}
			return cl.PrintObject(res)
		},
	}
	return cmd
}

// QueryAuthParamsCmd to query the current auth params
func QueryAuthParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current auth parameters",
		Args:  cobra.NoArgs,
		Long: strings.TrimSpace(`Query the current auth parameters:

$ resolute query auth params
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			cq := query.Query{Client: cl, Options: query.DefaultOptions()}

			res, err := cq.QueryAuthParams()
			if err != nil {
				return err
			}
			return cl.PrintObject(res)
		},
	}
	return cmd
}
