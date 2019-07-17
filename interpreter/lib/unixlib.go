package lib

var unixLib = []byte(`
library_version = 1

def user_list(mount):
  out = []
  pwd = mount.cat("/etc/passwd")
  for line in pwd.splitlines():
    spl = line.split(':')
    if len(spl) > 4:
      out.append(struct(
        user=spl[0],
        pass_info=spl[1],
        uid=int(spl[2]),
        gid=int(spl[3]),
        info=spl[4],
      ))
  return out

def users(mount):
  return {x.user: x for x in user_list(mount)}

def configure_hostname(mount, hostname):
  mount.write('/etc/hostname', str(hostname).strip() + '\n', fs.perms.default)

def set_shadow_password(mount, user, password):
  shadow_data = mount.cat("/etc/shadow")
  shadow_perm = mount.stat("/etc/shadow").mode
  new_shadow_data = ''

  for line in shadow_data.splitlines(True):
    if line.startswith(user + ':'):
      spl = line.split(':')
      new_shadow_data += spl[0] + ':' + crypt.unix_hash(password) + ':' + ':'.join(spl[2:])
    else:
      new_shadow_data += line
  mount.write('/etc/shadow', new_shadow_data, shadow_perm)
`)
