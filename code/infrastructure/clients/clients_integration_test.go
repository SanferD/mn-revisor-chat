package clients

import (
	"context"
	"testing"

	"code/application"

	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	ctx := context.Background()

	httpClientHelper, err := InitializeHTTPClientHelper()
	assert.NoError(t, err, "error on InitializeHTTPClientHelper: %v", err)

	t.Run("test GetHTML", func(t *testing.T) {
		output, err := httpClientHelper.GetHTML(ctx, application.MNRevisorStatutesURL)
		assert.NoError(t, err, "error on GetHTML: %v", err)
		assert.NotEmpty(t, output, "html page is empty: %v", output)
	})
}
