package vlt

import (
	"reflect"
	"testing"
)

func TestGetSecret(t *testing.T) {
	secret, err := client.GetSecret("TEST_SECRET1")
	if err != nil {
		t.Errorf("client GetSecret error: %s", err.Error())
	}

	if secret.Name != "TEST_SECRET1" {
		t.Errorf("client GetSecret error: Could not get named secret TEST_SECRET1")
	}
}

func TestCreateSecret(t *testing.T) {
	newSecret, err := client.CreateSecret("TEST_SECRET3", "created_secret")
	if err != nil {
		t.Errorf("client CreateSecret error: %s", err.Error())
	}

	if newSecret.LatestVersion < 1 {
		t.Errorf("CreateSecret did not make a new secret")
	}
}

func TestListSecrets(t *testing.T) {
	secrets, err := client.ListSecrets()
	if err != nil {
		t.Errorf("client ListSecrets error: %s", err.Error())
	}

	expected := []string{"TEST_SECRET1", "TEST_SECRET2", "TEST_SECRET3"}
	if !reflect.DeepEqual(secrets, expected) {
		t.Errorf("ListSecrets did not result in the expected value")
	}
}

func TestGetAllSecrets(t *testing.T) {
	secrets, err := client.GetAllSecrets()
	if err != nil {
		t.Errorf("client GetAllSecrets error: %s", err.Error())
	}

	expectedResult := []map[string]string{
		{"TEST_SECRET1": "this_is_a_value"},
		{"TEST_SECRET2": "this_is_a_value_too"},
		{"TEST_SECRET3": "created_secret"},
	}

	for idx, secret := range secrets {
		testCombo := map[string]string{secret.Name: secret.Version.Value}
		if !reflect.DeepEqual(testCombo, expectedResult[idx]) {
			t.Errorf("GetAllSecrets result did not match expected")
		}
	}

}

func TestDeleteSecret(t *testing.T) {
	_, err := client.CreateSecret("TEST_SECRET4", "delete_me")
	if err != nil {
		t.Errorf("DeleteSecret test could not create a secret to delete: %s", err.Error())
	}

	err = client.DeleteSecret("TEST_SECRET4")
	if err != nil {
		t.Errorf("DeleteSecret did not successfully delete secret: %s", err.Error())
	}
}
