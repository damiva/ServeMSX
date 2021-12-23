# ServeMSX
DIY content server for [MediaStation X](https://msx.benzac.de/info/)<br>**It is under development, for testing purposes only**

## Functionality
### Now:
- Serves local video & audio files
- Serves local torrent files ([TorrServer](https://github.com/YouROK/TorrServer/releases) need to be used)
- Serves TorrServer's torrents ([TorrServer](https://github.com/YouROK/TorrServer/releases) need to be used)
- Serves plugins, written on [Tengo language](https://github.com/d5/tengo) (see [Plugins Development Manual](PLUGINS.md))
- One plugin is dveloped for example: [tivix](https://github.com/damiva/ServeMSX-Plugs)
### In development:
- Plugins installation (and updates) automation
- Self update automation
### To do:
- More plugins
## Installation
Choose the apropriate file for your OS/Architecture from the releases, download it and just run.<br>It can be also installed as a service:
- For **Windows**, please use [NSSM](https://nssm.cc/usage).
- For **Linux** use native service manager, for example, for Systemd you can:
  1. use the file: [ServeMSX.service](ServeMSX.service) 
  2. put it to **/etc/systemd/system/**
  3. run command: <pre># systemctl enable ServeMSX && systemctl start ServeMSX</pre>
- For **OS X** use native service manager Launchd, for example, you can:
  1. use the file: [damia.ServeMSX.daemon.plist](damia.ServeMSX.daemon.plist)
  2. put it to **/Library/LaunchDaemons/** 
  3. run command: <pre># launchctl load /Library/LaunchDaemons/damiva.ServeMSX.daemon.plist</pre>
### Run paramters:
**ServeMSX [options]**<br>Where **[options]** can be one or more of:
- **[IP]<:PORT>** - the address of the http server is listen to (default is **:8008**)
- **-i** - do not log info messages (recomended to reduce log size)
- **-t** - do not print timestamp in logs (useful for systemd service manager)
- **-s** - skip verifying TLS sertificates (useful for tiny OS, like on routers)
### Note for running as service:
- Errors logs to STDERR, 
- Info messages logs to STDOUT,
- It should be restarted on successful (code 0) exit, becuse it exits succesfully only when it is restarting (manually from MSX or for self updating)
## Setup
### Media Station X
[Install MediaStation X on your TV](https://msx.benzac.de/info/?tab=PlatformSupport), run it, go to **Settings -> Start Parameter -> Setup** and enter the address (default port is 8008) of the machine where ServeMSX is running.
### Local media files
In the working directory of ServeMSX, create symbolic links with the following name to your folder:
- for video files: **video**
- for music files: **music**
### Torrents
To play torrents online, you should install and use [TorrServer](https://github.com/YouROK/TorrServer/releases). In the ServeMSX on Media Station X goto **Settings -> TorrServer** and enter the address (default port is 8090) of the machine where TorrServer is running (if it is the same maching with ServeMSX, it will be detected automatically).
### Plugins
For now, installation of the plugins is manual:
- download the plugins (mentioned above)
- in the working dir of ServeMSX create the folder **plugins** and put there the plugins (e.g. **plugins/tivix/**)
