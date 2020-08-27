package actions

import (
	"golang.org/x/crypto/ssh"
	"microtools-gossh/app/types"
	"microtools-gossh/app/utils"
)

// Connect to server
func Connect(option types.SshOption) (client *ssh.Client, err error) {
	auth, err := Auth(option)
	if err != nil {
		return
	}
	config := ssh.ClientConfig{
		User:            option.Username,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	addr := utils.GetAddr(option.Host, uint(option.Port))
	client, err = ssh.Dial("tcp", addr, &config)
	return
}
