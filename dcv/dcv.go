// Package dcv contains common operations needed to administrate DCV
package dcv

import (
	"encoding/json"
	"os/exec"
	"slices"
)

type VirtualSession struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Owner string `json:"owner"`
	User  string `json:"user"`
}

func CreateVirtualSessionFromUsername(username string) error {
	v := VirtualSession{
		ID:    username,
		Name:  username,
		Owner: username,
		User:  username,
	}
	return createVirtualSession(&v)
}

// PruneVirtualSessions is closing all sessions for which no owner
// can be found.
func PruneVirtualSessions(users []string) error {
	sessions, err := listVirtualSessions()
	if err != nil {
		return err
	}
	for _, s := range sessions {
		if slices.Contains(users, s.Owner) {
			continue
		}
		// user doesn't exist but session does
		err := deleteVirtualSession(s.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func deleteVirtualSession(id string) error {
	cmd := exec.Command(
		"dcv",
		"close-session",
		id,
	)
	return cmd.Run()
}

func createVirtualSession(s *VirtualSession) error {
	isCreated, err := isVirtualSessionCreated(s.ID)
	if err != nil {
		return err
	}
	if isCreated {
		return nil
	}
	cmd := exec.Command(
		"dcv",
		"create-session",
		"--type", "virtual",
		"--name", s.Name,
		"--user", s.User,
		"--owner", s.Owner,
		s.ID,
	)
	return cmd.Run()
}

func listVirtualSessions() ([]VirtualSession, error) {
	var sessions []VirtualSession
	cmd := exec.Command(
		"dcv",
		"list-sessions",
		"--type", "virtual",
		"--json",
	)
	data, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &sessions)
	return sessions, err
}

func isVirtualSessionCreated(id string) (bool, error) {
	sessions, err := listVirtualSessions()
	if err != nil {
		return true, err
	}
	for _, s := range sessions {
		if s.ID == id {
			return true, nil
		}
	}
	return false, nil
}
