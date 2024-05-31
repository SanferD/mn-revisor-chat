package comms

import (
	"code/infrastructure/settings"
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	msg                = "hello from comms-test.go"
	phoneNumberEnvName = "MY_PHONE_NUMBER"
)

func TestCommsTest(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)

	mySettings, err := settings.GetSettings()
	assert.NoError(err, "error on get settings: %v", err)
	commsHelper, err := InitializeSinchHelper(ctx, mySettings.SinchAPIToken, mySettings.SinchProjectID, mySettings.SinchVirtualPhoneNumber, mySettings.ContextTimeout)
	assert.NoError(err, "error on initializing sinch helper: %v", err)

	t.Run("test send sms", func(t *testing.T) {
		phoneNumber := os.Getenv(phoneNumberEnvName)
		err = commsHelper.SendMessage(ctx, phoneNumber, msg)
		assert.NoError(err, "error on send message: %v", err)
	})
}
