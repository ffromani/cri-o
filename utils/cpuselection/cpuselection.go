package cpuselection

import (
	"fmt"
	"os/exec"

	specs "github.com/opencontainers/runtime-spec/specs-go"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"
)

const (
	annotationHousekeepingCPUs = "housekeeping-cpus.crio.io"
	// cri-o depends on systemd, which in turns requires util-linux, which contain taskset. So we can assume it's avaialble.
	tasksetPath = "/usr/bin/taskset"
)

func PinnedCommand(cpus cpuset.CPUSet, name string, args ...string) *exec.Cmd {
	if cpus.Size() == 0 {
		return exec.Command(name, args...)
	}
	pinnedName, pinnedArgs := PinnedCommandline(cpus, name, args...)
	return exec.Command(pinnedName, pinnedArgs...)
}

func PinnedCommandline(cpus cpuset.CPUSet, name string, args ...string) (string, []string) {
	initialArgs := []string{
		"-c",
		cpus.String(),
		name,
	}
	return tasksetPath, append(initialArgs, args...)
}

func GetHousekeepingCPUSet(containerID string, annotations fields.Set, spec specs.Spec) (cpuset.CPUSet, error) {
	_, ok := annotations[annotationHousekeepingCPUs]
	if !ok {
		return cpuset.CPUSet{}, nil
	}
	cpus, err := GetContainerCPUSet(containerID, spec)
	if err != nil {
		return cpuset.CPUSet{}, err
	}
	cpuIDs := cpus.ToSlice()
	return cpuset.NewCPUSet(cpuIDs[0]), nil
}

func GetContainerCPUSet(containerID string, spec specs.Spec) (cpuset.CPUSet, error) {
	lspec := spec.Linux
	if lspec == nil ||
		lspec.Resources == nil ||
		lspec.Resources.CPU == nil ||
		lspec.Resources.CPU.Cpus == "" {
		return cpuset.CPUSet{}, fmt.Errorf("failed to find the container %q CPUs", containerID)
	}
	return cpuset.Parse(lspec.Resources.CPU.Cpus)
}
