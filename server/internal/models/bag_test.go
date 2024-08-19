package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/the-clothing-loop/website/server/internal/models"
)

func TestBagAddUpdatedUser(t *testing.T) {
	f := func(name, initial, expected string) {
		t.Helper()
		b := &models.Bag{}
		b.LastUserEmailToUpdate = initial
		b.AddLastUserEmailToUpdateFifo("new@example.com")
		assert.Equal(t, b.LastUserEmailToUpdate, expected, name)
	}

	f("Add to empty", "", "new@example.com")
	f("Add to len 1", "old@example.com", "old@example.com,new@example.com")
	f("Add to len 2", "old2@example.com,old@example.com", "old2@example.com,old@example.com,new@example.com")
	f("Add to len 3", "old3@example.com,old2@example.com,old@example.com", "old3@example.com,old2@example.com,old@example.com,new@example.com")
	f("Add to len 4", "old4@example.com,old3@example.com,old2@example.com,old@example.com", "old3@example.com,old2@example.com,old@example.com,new@example.com")
	f("Add to len 5", "old5@example.com,old4@example.com,old3@example.com,old2@example.com,old@example.com", "old3@example.com,old2@example.com,old@example.com,new@example.com")

	f("Add duplicate", "new@example.com", "new@example.com")
	f("Add duplicate with old", "old@example.com,new@example.com", "old@example.com,new@example.com")
	f("Add duplicate with other edit", "new@example.com,old@example.com", "new@example.com,old@example.com,new@example.com")
}
