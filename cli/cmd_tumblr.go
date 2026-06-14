package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func (a *App) tagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag <tag>",
		Short: "List posts tagged with a given tag",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			posts, err := a.client.Tagged(cmd.Context(), args[0], a.limit)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(posts, len(posts))
		},
	}
	return cmd
}

func (a *App) postsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "posts <blog>",
		Short: "List posts from a blog",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			posts, err := a.client.Posts(cmd.Context(), args[0], a.limit)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(posts, len(posts))
		},
	}
	return cmd
}

func (a *App) infoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info <blog>",
		Short: "Show blog metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			blog, err := a.client.BlogInfo(cmd.Context(), args[0])
			if err != nil {
				return mapFetchErr(err)
			}
			if blog == nil {
				_, _ = fmt.Fprintln(os.Stderr, "blog not found")
				return codeError(exitNoData, nil)
			}
			return a.render([]*struct {
				Name        string `json:"name"        csv:"name"        tsv:"name"`
				Title       string `json:"title"       csv:"title"       tsv:"title"`
				Description string `json:"description" csv:"description" tsv:"description"`
				Posts       int    `json:"posts"       csv:"posts"       tsv:"posts"`
				Updated     int64  `json:"updated"     csv:"updated"     tsv:"updated"`
				URL         string `json:"url"         csv:"url"         tsv:"url"`
			}{{
				Name:        blog.Name,
				Title:       blog.Title,
				Description: blog.Description,
				Posts:       blog.Posts,
				Updated:     blog.Updated,
				URL:         blog.URL,
			}})
		},
	}
	return cmd
}
