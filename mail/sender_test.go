package mail

import (
	"testing"

	"github.com/ahmedkhaeld/simplebank/util"
	"github.com/stretchr/testify/require"
)

func TestSendEmailWithGmail(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	config, err := util.LoadEnv("..")
	require.NoError(t, err)

	sender := NewGmail(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)

	subject := "A test email"
	content := `
	<h1>Hello world</h1>
	<p>This is a test message from <a href="http://techschool.guru">Tech School</a></p>
	`
	to := []string{"ahmed.alsaeidy.711@gmail.com"}
	attachFiles := []string{"../README.md"}

	err = sender.Send(subject, content, to, nil, nil, attachFiles)
	require.NoError(t, err)
}