package environment

import (
	"os"
	"testing"
)

func TestSystemEnv(t *testing.T) {
	env := NewSystemEnv()

	// Тест получения текущей директории
	dir, err := env.CurrentDir()
	if err != nil {
		t.Errorf("CurrentDir() failed: %v", err)
	}
	if dir == "" {
		t.Error("CurrentDir() returned empty string")
	}

	// Тест смены директории
	originalDir := dir
	tempDir := os.TempDir()
	
	err = env.ChangeDir(tempDir)
	if err != nil {
		t.Errorf("ChangeDir() failed: %v", err)
	}

	// Возвращаемся обратно
	defer env.ChangeDir(originalDir)

	// Проверяем, что директория изменилась
	newDir, _ := env.CurrentDir()
	if newDir == originalDir {
		t.Error("Directory did not change")
	}

	// Тест переменных окружения
	testVar := "TEST_SHELL_VAR"
	testValue := "test_value"
	
	os.Setenv(testVar, testValue)
	defer os.Unsetenv(testVar)

	value := env.Variable(testVar)
	if value != testValue {
		t.Errorf("Variable() = %q, want %q", value, testValue)
	}

	// Тест домашней директории
	home, err := env.HomeDir()
	if err != nil {
		t.Errorf("HomeDir() failed: %v", err)
	}
	if home == "" {
		t.Error("HomeDir() returned empty string")
	}
}