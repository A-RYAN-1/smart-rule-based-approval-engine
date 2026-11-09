// Command ruledraft-eval measures the rule-drafting prompt against the
// checked-in evaluation set using the real model and prints the exact-match
// result. It needs GEMINI_API_KEY and spends free-tier calls: one per case.
//
//	GEMINI_API_KEY=... go run ./cmd/ruledraft-eval [-model gemini-2.5-flash-lite] [-out result.json]
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ankita-advitot/rule_based_approval_engine/app/rule_drafting"
)

func main() {
	model := flag.String("model", "gemini-2.5-flash-lite", "Gemini model name")
	out := flag.String("out", "", "write the full report as JSON to this file")
	pause := flag.Duration("pause", 5*time.Second, "wait between calls to stay inside the free-tier rate limit")
	flag.Parse()

	key := os.Getenv("GEMINI_API_KEY")
	if key == "" {
		fmt.Fprintln(os.Stderr, "GEMINI_API_KEY is not set")
		os.Exit(2)
	}
	set, err := rule_drafting.LoadEvalSet()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	provider := &pausing{inner: rule_drafting.NewGemini(key, *model), gap: *pause}
	rep, err := rule_drafting.Evaluate(context.Background(), provider, set)
	if err != nil {
		fmt.Fprintln(os.Stderr, "stopped early:", err)
	}

	for _, r := range rep.Results {
		if !r.Pass {
			fmt.Printf("FAIL %q\n     want %s\n     got  %s\n", r.Sentence, r.Want, r.Got)
		}
	}
	date := time.Now().UTC().Format("2006-01-02")
	fmt.Printf("\nprompt %s, model %s, %s\n", rep.PromptVersion, *model, date)
	fmt.Printf("exact match %d/%d\n", rep.ExactMatch, rep.Total)
	fmt.Printf("rule cases  %d/%d\n", rep.RulesMatched, rep.Rules)
	fmt.Printf("refusals    %d/%d\n", rep.RejectsKept, rep.Rejects)

	if *out != "" {
		data, _ := json.MarshalIndent(map[string]any{"model": *model, "date": date, "report": rep}, "", " ")
		if werr := os.WriteFile(*out, data, 0o644); werr != nil {
			fmt.Fprintln(os.Stderr, werr)
			os.Exit(1)
		}
	}
	if err != nil {
		os.Exit(1)
	}
}

// pausing spaces the calls out so the free-tier rate limit is not hit.
type pausing struct {
	inner rule_drafting.Provider
	gap   time.Duration
}

func (p *pausing) Complete(ctx context.Context, system, user string) (string, error) {
	time.Sleep(p.gap)
	return p.inner.Complete(ctx, system, user)
}
