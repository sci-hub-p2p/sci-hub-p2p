package cmd

import "github.com/spf13/cobra"

func MarkFlagsRequired(c *cobra.Command, flags ...string) error {
	for _, flag := range flags {
		err := c.MarkFlagRequired(flag)
		if err != nil {
			return err
		}
	}
	return nil
}
