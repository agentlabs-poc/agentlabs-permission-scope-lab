package cli

import (
	"agentlabs.local/abv/application"
	"agentlabs.local/abv/domain"
	"context"
	"errors"
	"io"
	"os"
)

const maxAssignmentBytes = 1 << 20

func dispatch(ctx context.Context, command string, positional []string, flags map[string]string, in io.Reader, out, diag io.Writer, api application.API, area domain.Area) int {
	switch command {
	case "inspect":
		if err := inspect(ctx, api, area, positional[0], positional[1], out); err != nil {
			return report(diag, err)
		}
		return 0
	case "check", "assign":
		raw, err := readInput(flags["--file"], in)
		if err != nil {
			return report(diag, err)
		}
		if command == "check" {
			err = check(ctx, api, area, raw, out)
		} else {
			err = assign(ctx, api, area, flags["--fixture-context"], raw, out, diag)
		}
		if err != nil {
			return report(diag, err)
		}
		return 0
	}
	return report(diag, domain.ErrUnsupported)
}

func readInput(path string, in io.Reader) (raw []byte, err error) {
	reader := in
	var file *os.File
	if path != "-" {
		file, err = os.Open(path)
		if err != nil {
			return nil, errors.Join(domain.ErrUnavailable, err)
		}
		defer func() { err = errors.Join(err, file.Close()) }()
		reader = file
	}
	raw, err = io.ReadAll(io.LimitReader(reader, maxAssignmentBytes+1))
	if err != nil {
		return nil, errors.Join(domain.ErrUnavailable, err)
	}
	if len(raw) > maxAssignmentBytes {
		return nil, domain.ErrMalformed
	}
	return raw, nil
}
