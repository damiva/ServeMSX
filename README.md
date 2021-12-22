# ServeMSX
DIY content server for MediaStation X<br>**It is under development, for testing purposes only**

## Functionality
### Now:
- Serves local (server) video & audio files
- Serves local toeernt files ([TorrServer](https://github.com/YouROK/TorrServer/releases) need to be installed)
- Serves TorrSerever's torrents ([TorrServer](https://github.com/YouROK/TorrServer/releases) need to be installed)
- Serves plugins, written on [Tengo language](https://github.com/d5/tengo) (see [Plugins Development Manual](PLUGINS.md))
- One plugin is dveloped for example: [tivix](https://github.com/damiva/ServeMSX-Plugs)
### In development:
- Plugins installation (and updates) automation
- Self update automation
### To do:
- More plugins
## Installation
Choose the apropriate file for your OS/Architecture from the releases, download it and just run.<br>It can be also installed as a service:
- For windows, please use [NSSM](https://nssm.cc/usage).
- For Linux/OSX use the native service manager (e.g. systemd, launchd, etc).
### Run paramters:
**ServeMSX [options]**<br>Where **[options]** can be one or more of:
- **[IP]<:PORT>** - the address of the http server is listen to (default is **:8008**)
- **-i** - do not log info messages (recomended to reduce log size)
- **-t** - do not print timestamp in logs (useful for systemd service manager)
- **-s** - skip verifying TLS sertificates (useful for tiny OS, like on routers)
