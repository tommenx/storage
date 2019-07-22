package lvm

type Mounter interface {
	// Mount mounts source to target with the given fstype and options.
	Mount(source, target, fsType string, options ...string) error
	// Unmount unmounts the given target
	Unmount(target string) error
	// If the folder doesn't exist, it will call 'mkdir -p'
	EnsureFolder(target string) error
	// Format formats the source with the given filesystem type
	Format(source, fsType string) error
	// IsFormatted checks whether the source device is formatted or not. It
	// returns true if the source device is already formatted.
	IsFormatted(source string) (bool, error)

	// IsMounted checks whether the target path is a correct mount (i.e:
	// propagated). It returns true if it's mounted. An error is returned in
	// case of system errors or if it's mounted incorrectly.
	IsMounted(target string) (bool, error)
}
