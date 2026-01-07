package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
)

type InitResult struct {
	Handle       string
	Model        string
	Description  string
	IgnoreShared bool
	System       string
}

type InitForm struct{}

func NewInitForm() *InitForm { return &InitForm{} }

func (f *InitForm) Run(ctx context.Context, models []string) (InitResult, error) {
	if len(models) == 0 {
		return InitResult{}, fmt.Errorf("no models configured; update config provider models")
	}

	var res InitResult

	options := make([]huh.Option[string], 0, len(models))
	for _, m := range models {
		options = append(options, huh.NewOption(m, m))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Agent handle").
				Placeholder("@myagent").
				Value(&res.Handle).
				Validate(func(v string) error {
					if v == "" {
						return fmt.Errorf("required")
					}
					if !strings.HasPrefix(v, "@") {
						return fmt.Errorf("handle must start with @")
					}
					return nil
				}),
			huh.NewInput().
				Title("Description").
				Placeholder("optional").
				Value(&res.Description),
			huh.NewSelect[string]().
				Title("Model").
				Options(options...).
				Value(&res.Model),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Ignore shared system message?").
				Value(&res.IgnoreShared),
			huh.NewText().
				Title("System message").
				Placeholder("You are a helpful assistant...").
				Value(&res.System).
				CharLimit(0),
		),
	).WithTheme(huh.ThemeCharm())

	if err := form.RunWithContext(ctx); err != nil {
		return InitResult{}, err
	}

	return res, nil
}
