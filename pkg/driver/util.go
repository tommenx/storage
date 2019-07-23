package driver

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"os"
	"os/exec"
	"strings"
)

type Mounter interface {
	// Mount mounts source to target with the given fstype and options.
	Mount(source, target, fsType string, options ...string) error
	// Umount unmounts the given target
	Umount(target string) error
	// If the folder doesn't exist, it will call 'mkdir -p'
	EnsureFolder(target string) error
	// Format formats the source with the given filesystem type
	Format(source, fsType string) error
	// IsMounted checks whether the target path is a correct mount (i.e:
	// propagated). It returns true if it's mounted. An error is returned in
	// case of system errors or if it's mounted incorrectly.
	IsMounted(target string) (bool, error)
}

type mounter struct {
	cmdPrefix string
}

func NewMounter(prefix string) Mounter {
	return &mounter{
		cmdPrefix: prefix,
	}
}

func (m *mounter) EnsureFolder(target string) error {
	mkdirCmd := "mkdir"
	_, err := exec.LookPath(mkdirCmd)
	if err != nil {
		if err == exec.ErrNotFound {
			return errors.New("mkdir do not supported")
		}
		return err
	}
	args := []string{"-p", target}
	glog.Infof("command is %s %v", mkdirCmd, args)
	_, err = exec.Command(mkdirCmd, args...).CombinedOutput()
	if err != nil {
		glog.Errorf("mkdir error, err=%+v", err)
		return err
	}
	return nil
}

func (m *mounter) Format(source, fsType string) error {
	mkfsCmd := fmt.Sprintf("mkfs.%s", fsType)
	_, err := exec.LookPath(mkfsCmd)
	if err != nil {
		if err == exec.ErrNotFound {
			return errors.New("fs type do not supported")
		}
		return err
	}
	if len(fsType) == 0 || len(source) == 0 {
		glog.Errorf("fs type or source path is not specified")
		return errors.New("fs type or source path is not specified")
	}
	args := []string{}
	if fsType == "ext4" || fsType == "ext3" {
		args = append(args, "-F")
	}
	args = append(args, source)
	glog.Infof("command is %s %v", mkfsCmd, args)
	out, err := exec.Command(mkfsCmd, args...).CombinedOutput()
	if err != nil {
		glog.Errorf("mkfs error, error is %+v, output is %s", err, string(out))
		return fmt.Errorf("error is %+v, output is %s", err, string(out))
	}
	return nil
}

func (m *mounter) Mount(source, target, fsType string, opts ...string) error {
	mountCmd := "mount"
	if len(m.cmdPrefix) != 0 {
		mountCmd = fmt.Sprintf("%s %s", m.cmdPrefix, "mount")
	}
	args := []string{}
	if fsType == "" {
		return errors.New("fs type is not specified for mounting the volume")
	}

	if source == "" {
		return errors.New("source is not specified for mounting the volume")
	}

	if target == "" {
		return errors.New("target is not specified for mounting the volume")
	}

	args = append(args, "-t", fsType)
	if len(opts) > 0 {
		args = append(args, "-o", strings.Join(opts, ","))
	}
	args = append(args, source)
	args = append(args, target)
	err := os.MkdirAll(target, 0750)
	if err != nil {
		return err
	}
	glog.Infof("mount command is %s %v", mountCmd, args)
	out, err := exec.Command(mountCmd, args...).CombinedOutput()
	if err != nil {
		glog.Errorf("mount failed, err=%+v, output is %s", err, string(out))
		return err
	}
	return nil
}

func (m *mounter) Umount(target string) error {
	umountCmd := "umount"
	if len(m.cmdPrefix) != 0 {
		umountCmd = fmt.Sprintf("%s %s", m.cmdPrefix, "umount")
	}
	if len(target) == 0 {
		return errors.New("target is not specified for unmounting the volume")
	}
	args := []string{target}
	glog.Infof("umount command is %s %v", umountCmd, args)
	out, err := exec.Command(umountCmd, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("unmounting failed: %v cmd: '%s %s' output: %q",
			err, umountCmd, target, string(out))
	}
	return nil
}

func (m *mounter) IsMounted(target string) (bool, error) {
	if target == "" {
		return false, errors.New("target is not specified for checking the mount")
	}
	findmntCmd := "grep"
	findmntArgs := []string{target, "/proc/mounts"}
	out, err := exec.Command(findmntCmd, findmntArgs...).CombinedOutput()
	outStr := strings.TrimSpace(string(out))
	if err != nil {
		if outStr == "" {
			return false, nil
		}
		return false, fmt.Errorf("checking mounted failed: %v cmd: %q output: %q",
			err, findmntCmd, outStr)
	}
	if strings.Contains(outStr, target) {
		return true, nil
	}
	return false, nil
}

func Run(cmd string) (string, error) {
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Failed to run cmd: " + cmd + ", with out: " + string(out) + ", with error: " + err.Error())
	}
	return string(out), nil
}

func GetCmd(cmd string, args []string) string {
	str := cmd
	for _, v := range args {
		str += " " + v
	}
	return str
}
