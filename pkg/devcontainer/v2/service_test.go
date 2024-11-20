package devcontainer_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hide-org/hide/pkg/daytona"
	"github.com/hide-org/hide/pkg/devcontainer/v2"
)

func TestDaytonaService_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		config := daytona.NewConfiguration()
		api := daytona.NewAPIClient(config)
		service := devcontainer.NewDaytonaService(api)

		container, err := service.Create(context.Background(), "https://github.com/hide-org/hide")
		assert.NoError(t, err)
		assert.NotNil(t, container)
	})
}
