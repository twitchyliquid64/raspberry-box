package lib

var piLib = []byte(`
load("unix.lib", "configure_hostname", "set_shadow_password")
library_version = 2

def assert_valid_partitions(parts):
  if len(parts) < 3:
    crash("expected >=3 partitions, got " + str(len(parts)))

  if parts[0].type_name != "FAT32-LBA":
    crash("expected first partition to be FAT32-LBA, got " + parts[0].type_name)
  if parts[1].type_name != "Native Linux":
    crash("expected second partition to be Native Linux, got " + parts[1].type_name)

def load_img(img, min_mb=None):
  do_resize = False
  if min_mb:
    cur_size = fs.stat(img).size
    if cur_size < (min_mb * 1024 * 1024):
      print('Existing image is too small (%sMb < %sMb), resizing' % (str(cur_size // 1024 // 1024), str(min_mb)))
      fs.truncate(img, min_mb * 1024 * 1024)
      fs.expand_partition(img, partition=2)
      do_resize = True

  partitions = fs.read_partitions(img)
  assert_valid_partitions(partitions)
  ext4 = fs.mnt_ext4(img, partitions[1], do_resize)
  fat = fs.mnt_vfat(img, partitions[0])
  return struct(ext4=ext4,fat=fat)



def configure_pi_hostname(image, hostname):
  configure_hostname(image.ext4, hostname)

def enable_ssh(image):
  image.fat.write('ssh', '', fs.perms.default)

def cmdline(image):
  return image.fat.cat("/cmdline.txt").strip()

def disable_resize(image):
  image.fat.write("/cmdline.txt", cmdline(image).replace('init=/usr/lib/raspi-config/init_resize.sh', ''), fs.perms.default)
  if image.ext4.exists('etc/init.d/resize2fs_once'):
    image.ext4.remove('etc/init.d/resize2fs_once')
  if image.ext4.exists('ext4_mnt/etc/rc3.d/S01resize2fs_once'):
    image.ext4.remove('ext4_mnt/etc/rc3.d/S01resize2fs_once')

def configure_pi_password(image, password):
  set_shadow_password(image.ext4, "pi", password)



def configure_wifi_network(image, ssid, password):
  confMode = image.ext4.stat('/etc/wpa_supplicant/wpa_supplicant.conf').mode
  c = net.wifi.SupplicantConfig(
  	control_interface='/run/wpa_supplicant',
  	allow_update_config = True,
  	country_code='US',
  	device_name='wlan0',
  	networks=[net.wifi.Network(
    	ssid = ssid,
    	psk = password,
    )],
  )
  image.ext4.write('/etc/wpa_supplicant/wpa_supplicant.conf', str(c), confMode)

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




def run_on_boot(image, name, cmdLine, user='', group=''):
  if not name.endswith('.service'):
    name += '.service'
  s = systemd.Service(
  	exec_start=cmdLine,
  	restart=systemd.const.restart_never,
    user=user,
    group=group,
    stdout=systemd.out.console + systemd.out.journal,
    stderr=systemd.out.console + systemd.out.journal,
  )
  s.kill_mode = systemd.const.killmode_controlgroup
  s.ignore_sigpipe = True

  systemd.install(image.ext4, name, systemd.Unit(
    service=s,
  	description="run on startup",
  	after=['basic.target'],
    required_by=['multi-user.target'],
  ))
  if not systemd.is_enabled_on_target(image.ext4, name, 'multi-user.target'):
    systemd.enable_target(image.ext4, name, 'multi-user.target')

pi = struct(
  library_version=library_version,
  assert_valid_partitions=assert_valid_partitions,
  load_img=load_img,
  configure_static_ethernet=configure_static_ethernet,
  configure_dynamic_ethernet=configure_dynamic_ethernet,
  configure_hostname=configure_pi_hostname,
  enable_ssh=enable_ssh,
  cmdline=cmdline,
  disable_resize=disable_resize,
  configure_pi_password=configure_pi_password,
  configure_wifi_network=configure_wifi_network,
  run_on_boot=run_on_boot,
)`)
