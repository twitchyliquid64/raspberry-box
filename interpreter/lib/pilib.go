package lib

var piLib = []byte(`
library_version = 1

def assert_valid_partitions(parts):
  if len(parts) < 3:
    crash("expected >=3 partitions, got " + str(len(parts)))

  if parts[0].type_name != "FAT32-LBA":
    crash("expected first partition to be FAT32-LBA, got " + parts[0].type_name)
  if parts[1].type_name != "Native Linux":
    crash("expected second partition to be Native Linux, got " + parts[1].type_name)

pi = struct(
  library_version=library_version,
  assert_valid_partitions=assert_valid_partitions,
)`)
