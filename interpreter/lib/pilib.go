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

def load_img(img):
  partitions = fs.read_partitions(img)
  assert_valid_partitions(partitions)
  ext4 = fs.mnt_ext4(img, partitions[1])
  fat = fs.mnt_vfat(img, partitions[0])
  return struct(ext4=ext4,fat=fat)

def configure_static_ethernet(image, address=None, router=None, dns='8.8.8.8'):
  static = net.StaticProfile(
    interface='eth0',
    network=address,
    routers=[router],
    dns=[dns],
  )
  config = net.DHCPClient(profiles = [static])
  image.ext4.write('/etc/dhcpcd.conf', str(config), fs.perms.default)

def configure_dynamic_ethernet(image, lease_seconds=60*60*12, hostname=None):
  dynamic = net.DHCPProfile(
    interface='eth0',
    lease_seconds=lease_seconds,
  )
  if hostname:
    dynamic.set_hostname(hostname)
  config = net.DHCPClient(profiles = [dynamic])
  image.ext4.write('/etc/dhcpcd.conf', str(config), fs.perms.default)

def configure_hostname(image, hostname):
  image.ext4.write('/etc/hostname', str(hostname).strip() + '\n', fs.perms.default)

def enable_ssh(image):
  image.fat.write('ssh', '', fs.perms.default)

def cmdline(image):
  return image.fat.cat("/cmdline.txt").strip()

def disable_resize(image):
  image.fat.write("/cmdline.txt", cmdline(image).replace('init=/usr/lib/raspi-config/init_resize.sh', ''), fs.perms.default)
  if image.ext4.exists('etc/init.d/resize2fs_once'):
    print('Deleting: %s' % 'etc/init.d/resize2fs_once')
    image.ext4.remove('etc/init.d/resize2fs_once')
  if image.ext4.exists('ext4_mnt/etc/rc3.d/S01resize2fs_once'):
    print('Deleting: %s' % 'ext4_mnt/etc/rc3.d/S01resize2fs_once')
    image.ext4.remove('ext4_mnt/etc/rc3.d/S01resize2fs_once')

pi = struct(
  library_version=library_version,
  assert_valid_partitions=assert_valid_partitions,
  load_img=load_img,
  configure_static_ethernet=configure_static_ethernet,
  configure_dynamic_ethernet=configure_dynamic_ethernet,
  configure_hostname=configure_hostname,
  enable_ssh=enable_ssh,
  cmdline=cmdline,
  disable_resize=disable_resize,
)`)
