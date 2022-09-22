package main

import (
	"bytes"
	"fmt"
	"github.com/hstreamdb/dev-deploy/embed"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"text/template"
)

func newInit() *cobra.Command {
	cmd := &cobra.Command{
		Use: "init",
		RunE: func(cmd *cobra.Command, args []string) error {
			fp := filepath.Join("config", "config.yaml")
			tpl, err := embed.ReadConfig(fp)
			if err != nil {
				return fmt.Errorf("get config.yaml template error: %s\n", err.Error())
			}
			cfg, err := template.New("DefaultConfig").Parse(string(tpl))
			if err != nil {
				return fmt.Errorf("render config.yaml template error: %s\n", err.Error())
			}

			content := bytes.NewBufferString("")
			if err := cfg.Execute(content, nil); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), content.String())

			if err := os.MkdirAll("template/script", 0755); err != nil {
				return fmt.Errorf("create template config path error: %s\n", err.Error())
			}
			if err := os.MkdirAll("template/prometheus", 0755); err != nil {
				return fmt.Errorf("create template prometheus path error: %s\n", err.Error())
			}
			if err := os.MkdirAll("template/grafana/dashboards", 0755); err != nil {
				return fmt.Errorf("create template grafana path error: %s\n", err.Error())
			}
			if err := os.MkdirAll("template/grafana/datasources", 0755); err != nil {
				return fmt.Errorf("create template grafana path error: %s\n", err.Error())
			}
			if err := os.WriteFile("template/config.yaml", content.Bytes(), 0664); err != nil {
				return fmt.Errorf("write config file error: %s\n", err.Error())
			}
			return nil
		},
	}
	return cmd
}