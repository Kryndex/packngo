package packngo

import (
	"crypto/rand"
	"crypto/rsa"
	"reflect"
	"testing"

	"golang.org/x/crypto/ssh"
)

func makePubKey(t *testing.T) string {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("error generating test private key: %v", err)
	}

	pub, err := ssh.NewPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatalf("error generating test public key: %v", err)
	}
	return string(ssh.MarshalAuthorizedKey(pub))
}

func createKey(t *testing.T, c *Client, p string) *SSHKey {
	req := SSHKeyCreateRequest{
		Label:     "PACKNGO_TEST_KEY_DELETE_ME-" + randString8(),
		ProjectID: p,
		Key:       makePubKey(t),
	}

	key, _, err := c.SSHKeys.Create(&req)
	if err != nil {
		t.Fatalf("errored posting key: %v", err)
	}

	return key
}

func TestSSHKeyList(t *testing.T) {
	skipUnlessAcceptanceTestsAllowed(t)
	t.Parallel()
	c, _, teardown := setupWithProject(t)
	defer teardown()
	key := createKey(t, c, "")
	defer c.SSHKeys.Delete(key.ID)

	keys, _, err := c.SSHKeys.List()
	if err != nil {
		t.Fatalf("failed to get list of sshkeys: %v", err)
	}

	for _, k := range keys {
		if k.ID == key.ID {
			return
		}
	}
	t.Error("failed to find created key in list of keys retrieved")
}

func TestSSHKeyProjectList(t *testing.T) {
	skipUnlessAcceptanceTestsAllowed(t)
	t.Parallel()
	c, projectID, teardown := setupWithProject(t)
	defer teardown()

	key := createKey(t, c, projectID)
	defer c.SSHKeys.Delete(key.ID)

	keys, _, err := c.SSHKeys.ProjectList(projectID)
	if err != nil {
		t.Fatalf("failed to get list of project sshkeys: %v", err)
	}

	if len(keys) != 1 {
		t.Fatal("there should be exactly one key for the project")
	}

	for _, k := range keys {
		if k.ID == key.ID {
			return
		}
	}
	t.Error("failed to find created project key in list of project keys retrieved")
}

func TestSSHKeyGet(t *testing.T) {
	skipUnlessAcceptanceTestsAllowed(t)
	t.Parallel()
	c, projectID, teardown := setupWithProject(t)
	defer teardown()
	user := createKey(t, c, "")
	defer c.SSHKeys.Delete(user.ID)
	proj := createKey(t, c, projectID)

	for _, k := range []*SSHKey{user, proj} {
		got, _, err := c.SSHKeys.Get(k.ID)
		if err != nil {
			t.Fatalf("failed to retrieve created key")
		}

		if !reflect.DeepEqual(k, got) {
			t.Errorf("keys do not match, want: %v, got:%v", k, got)
		}
	}
}

func TestSSHKeyCreate(t *testing.T) {
	skipUnlessAcceptanceTestsAllowed(t)
	t.Parallel()
	c, projectID, teardown := setupWithProject(t)
	defer teardown()

	req := SSHKeyCreateRequest{
		Label:     "PACKNGO_TEST_KEY_DELETE_ME-" + randString8(),
		ProjectID: projectID,
		Key:       makePubKey(t),
	}

	key, _, err := c.SSHKeys.Create(&req)
	if err != nil {
		t.Fatalf("errored posting key: %v", err)
	}

	if key.Label != req.Label {
		t.Fatalf("returned key label does not match, want: %v, got: %v", req.Label, key.Label)
	}
	if key.Key != req.Key {
		t.Fatalf("returned key does not match, want: %v, got: %v", req.Key, key.Key)
	}
}

func TestSSHKeyUpdate(t *testing.T) {
	skipUnlessAcceptanceTestsAllowed(t)
	t.Parallel()
	c, projectID, teardown := setupWithProject(t)
	defer teardown()
	key := createKey(t, c, projectID)

	newKey := makePubKey(t)
	req := SSHKeyUpdateRequest{
		ID:    key.ID,
		Key:   newKey,
		Label: key.Label + "-updated",
	}
	got, _, err := c.SSHKeys.Update(&req)
	if err != nil {
		t.Fatalf("failed to update key: %v", err)
	}

	if reflect.DeepEqual(key, got) {
		t.Fatalf("expected keys to differ, got: %v", key)
	}

	if got.Label != key.Label+"-updated" {
		t.Fatalf("expected updated label, want: %s-updated, got: %s", key.Label, got.Label)
	}
	if got.Key != newKey {
		t.Fatalf("expected updated key string, want: %s, got: %s", newKey, got.Key)
	}
}

func TestSSHKeyDelete(t *testing.T) {
	skipUnlessAcceptanceTestsAllowed(t)
	t.Parallel()
	c, projectID, teardown := setupWithProject(t)
	defer teardown()
	key := createKey(t, c, projectID)

	_, err := c.SSHKeys.Delete(key.ID)
	if err != nil {
		t.Fatalf("unable to delete key: %v", err)
	}

	unexpected, _, err := c.SSHKeys.Get(key.ID)
	if err == nil {
		t.Fatalf("expected an error getting key, got: %v", unexpected)
	}
}
