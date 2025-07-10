package docker

import (
	"encoding/json"
	"fmt"
	"os/exec"
    "strings"
)

type inspectResult []struct {
	Mounts []struct {
		Destination string `json:"Destination"`
		Source      string `json:"Source"`
	} `json:"Mounts"`
}

func GetVolumePath(containerName string) (string, error) {
    cmd := exec.Command("docker", "inspect", containerName)
    output, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("docker inspect failed: %v", err)
    }

    var result []struct {
        Mounts []struct {
            Source      string `json:"Source"`
            Destination string `json:"Destination"`
        } `json:"Mounts"`
    }

    if err := json.Unmarshal(output, &result); err != nil {
        return "", fmt.Errorf("failed to parse docker inspect output: %v", err)
    }

    if len(result) == 0 {
        return "", fmt.Errorf("no inspect data for container %s", containerName)
    }

    mounts := result[0].Mounts
    if len(mounts) == 0 {
        return "", fmt.Errorf("no mounts found in container %s", containerName)
    }

    for _, mount := range mounts {
        if !strings.HasPrefix(mount.Source, "/opt/serversentinel") {
            return mount.Source, nil
        }
    }

    return "", fmt.Errorf("no valid mount source found for container %s", containerName)
}
