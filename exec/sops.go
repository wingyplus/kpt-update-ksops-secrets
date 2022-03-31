// Copyright 2022 Neutron Soutmun <neutron@neutron.in.th>
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/neutronth/kpt-update-ksops-secrets/config"
)

type SopsEncryptionInterface interface {
	Encrypt(input string, recipients ...config.UpdateKSopsRecipient) (output string, err error)
}

type sops struct{}

func NewSopsEncryption() SopsEncryptionInterface {
	return &sops{}
}

func (s *sops) Encrypt(input string, recipients ...config.UpdateKSopsRecipient) (output string, err error) {
	var execErr, execOut bytes.Buffer

	cmdOpts := []string{
		"--input-type=yaml",
		"--output-type=yaml",
		"--encrypted-regex=^(data|stringData)$",
		"--encrypt",
	}

	recipientsOpts := cmdRecipientsOptions(recipients...)
	cmdOpts = append(cmdOpts, recipientsOpts...)
	cmdOpts = append(cmdOpts, "/dev/stdin")
	cmd := exec.Command("sops", cmdOpts...)
	cmd.Stdin = strings.NewReader(input)
	cmd.Stdout = &execOut
	cmd.Stderr = &execErr

	if e := cmd.Run(); e != nil {
		return "", fmt.Errorf("Sops encryption error: %v\n%s", e, execErr.String())
	}

	return execOut.String(), nil
}

func cmdRecipientsOptions(recipients ...config.UpdateKSopsRecipient) (opts []string) {
	for _, r := range recipients {
		switch r.Type {
		case "age":
			opts = append(opts, fmt.Sprintf("--age=%s", r.Recipient))
		case "pgp":
			opts = append(opts, fmt.Sprintf("--pgp=%s", r.Recipient))
		}
	}

	return
}
