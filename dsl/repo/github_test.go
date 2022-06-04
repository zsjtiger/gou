package repo

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGithubContentPublic(t *testing.T) {
	repo, err := NewRepo("github.com/yaoapp/gou", map[string]interface{}{})
	if err != nil {
		t.Fatal(err)
	}
	content, err := repo.Content("/tests/app/app.yao")
	if err != nil {
		t.Fatal(err)
	}
	assert.Greater(t, len(content), 0)
	assert.Contains(t, string(content), "Pet Hospital")
}

func TestGithubDir(t *testing.T) {
	repo, err := NewRepo("github.com/yaoapp/gou", map[string]interface{}{})
	if err != nil {
		t.Fatal(err)
	}
	dirs, err := repo.Dir("/tests/app")
	if err != nil {
		t.Fatal(err)
	}

	assert.Greater(t, len(dirs), 0)
	assert.Contains(t, dirs, "/tests/app/workshop.yao")
}

func TestGithubContentPrivate(t *testing.T) {
	url := os.Getenv("GOU_TEST_GITHUB_REPO")
	token := os.Getenv("GOU_TEST_GITHUB_TOKEN")
	repo, err := NewRepo(url, map[string]interface{}{"token": token})
	if err != nil {
		t.Fatal(err)
	}

	content, err := repo.Content("/README.md")
	if err != nil {
		t.Fatal(err)
	}

	assert.Greater(t, len(content), 0)
	assert.Contains(t, string(content), "# workshop-tests-private")
}

func TestGithubContentFail(t *testing.T) {
	repo, err := NewRepo("github.com/yaoapp/gou", map[string]interface{}{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = repo.Content("/test/app/app.yao")
	assert.EqualError(t, err, "Github API Error: 404 Not Found")
}
