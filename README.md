# raspberry-box

A lightweight tool to customize a raspberry pi image.

 * Setup wireless networks to connect to
 * Setup static or DHCP configuration
 * Install services or scripts to run at boot
 * Enable SSH
 * Disable image resizing at boot
 * Set user/root passwords

 And much more!

 ## Quickstart

**Install rbox**

 ```shell
 go install github.com/twitchyliquid64/raspberry-box/rbox

 ```

**Write out a config**

Save a file like this into your project directory (or any directory).

```python
# mypi.box
load('pi.lib', "pi")

# setup is called before build().
def setup(img):
    image = pi.load_img(img)
    return struct(image=image)

# build is called to actually build the image.
# The return value of setup() is passed to build().
def build(setup):
    pi.configure_hostname(setup.image, 'my-pi')
    pi.enable_ssh(setup.image)
    pi.configure_static_ethernet(setup.image, address='192.168.1.5/24', router='192.168.1.1')
    pi.configure_pi_password(setup.image, password='whelp')
    pi.configure_wifi_network(setup.image, ssid='test',password='network')

    pi.run_on_boot(setup.image, 'custom-print', '/bin/echo Custom script stated yo!!!!!!!!!!!')

```

**Run rbox**

```shell
cp 2019-07-10-raspbian-buster-lite.img mypi.img # Make a copy for customization
sudo ./rbox --img mypi.img --script mypi.box # Actually customize the image
```
