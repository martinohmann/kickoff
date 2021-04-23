package testutil

import "os"

// Unsetenv set an environment variable and returns a func to restore it to
// its previous value.
//
// Usage example in tests:
//
//   defer testutil.Setenv(t, key, value)()
//
func Setenv(key, value string) func() {
	oldVal, restoreNeeded := os.LookupEnv(key)

	os.Setenv(key, value)

	return func() {
		if restoreNeeded {
			os.Setenv(key, oldVal)
		} else {
			os.Unsetenv(key)
		}
	}
}

// Unsetenv unset an environment variable and returns a func to restore it to
// its previous value.
//
// Usage example in tests:
//
//   defer testutil.Unsetenv(t, key)()
//
func Unsetenv(key string) func() {
	oldVal, restoreNeeded := os.LookupEnv(key)

	os.Unsetenv(key)

	return func() {
		if restoreNeeded {
			os.Setenv(key, oldVal)
		}
	}
}
