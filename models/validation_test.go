package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserValidation(t *testing.T) {
	testCases := []struct {
		name     string
		user     User
		expected bool
	}{
		{
			name: "Valid User",
			user: User{
				Username: "validuser",
				Email:    "valid@example.com",
				Password: "securePassword123!",
			},
			expected: true,
		},
		{
			name: "Invalid Email",
			user: User{
				Username: "validuser",
				Email:    "invalid-email",
				Password: "securePassword123!",
			},
			expected: false,
		},
		{
			name: "Short Password",
			user: User{
				Username: "validuser",
				Email:    "valid@example.com",
				Password: "short",
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.user.Validate()
			if tc.expected {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestPostValidation(t *testing.T) {
	testCases := []struct {
		name     string
		post     Post
		expected bool
	}{
		{
			name: "Valid Post",
			post: Post{
				UserID:  1,
				Title:   "Valid Post Title",
				Content: "This is a valid post content",
			},
			expected: true,
		},
		{
			name: "Invalid Post - Empty Title",
			post: Post{
				UserID:  1,
				Title:   "",
				Content: "Content",
			},
			expected: false,
		},
		{
			name: "Invalid Post - Empty Content",
			post: Post{
				UserID:  1,
				Title:   "Title",
				Content: "",
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.post.Validate()
			if tc.expected {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
