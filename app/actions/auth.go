package actions

import (
	"golang.org/x/crypto/ssh"
	"microtools-gossh/app/types"
)

// Verify authentication method
func Auth(option types.SshOption) (auth []ssh.AuthMethod, err error) {
	// Priority detection key method
	if len(option.Key) != 0 {
		var signer ssh.Signer
		if len(option.PassPhrase) == 0 {
			signer, err = ssh.ParsePrivateKey(option.Key)
			if err != nil {
				return
			}
		} else {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(
				option.Key,
				option.PassPhrase,
			)
			if err != nil {
				return
			}
		}
		auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
		return
	}
	// Use password
	if option.Password != "" {
		auth = []ssh.AuthMethod{
			ssh.Password(option.Password),
		}
	}
	return
}
