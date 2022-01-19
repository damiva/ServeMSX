# Developing plugins for ServeMSX
Plugins for the ServeMSX is a small scripts for generating/parsing contents to show it in [Media Station X](https://msx.benzac.de/info/).

Plugins can be written on the server side on [tengo langiage](https://github.com/d5/tengo), on the client (Media Station X) side on JavaScript. 
The interface between server and client is JSON API.

For client side plugins and JSON API of Media Station X see: [msx.benzac.de/wiki](https://msx.benzac.de/wiki).

This docuemnt describes the structure of a plugin and tengo libraty for the scripts on server side.
## Plugin structure
Plugin is a folder with as minimum one required file **manifest.json**. The name of the plugin is the name of the folder (**do not use name "msx", becuse it is used by the system**). It places into the folder **plugins** in the working directory of SeveMSX. There can be any files in the plugin's folder, including images, video, audio, scripts. Script files (with extension **.tengo** will be executed, all other files will be send to the client as is (if requested). The default script file is **main.tengo**, it is executed if the plugin's folder requested.

The example of plugin's folder named "myplugin":
- :open_file_folder: plugins
  - :open_file_folder: myplugin
    - :spiral_notepad: manifest.json
    - :spiral_notepad: start.tengo
    - :spiral_notepad: main.tengo
    - :framed_picture: logo.png
### manifest.json
It should contains the JSON object with the following properties (all properties are optional):
- "**Label**": {string} - the title of the plugin showed in the main menu (if omited plugin name (folder name) will be used),
- "**Image**": {string} - the URL to the logo of the plugin (showed in the main menu, if set the next property "icon" is ignored),
- "**Icon**": {string} - the icon of the plugin (showed in the main menu if the image is not set, all posible icons see [here](https://msx.benzac.de/wiki/index.php?title=Icons),
- "**URL**": {string} - the URL with started data (if omited the "*http://ip:port/pluginName/*" is using, it means **main.tengo** will be run),
- "**Torrent**" : {boolean} - indecates if the plugin depends on TorrServer is used (if omited means *false*).

In the URLs ("Image", "URL") you can use {BASE_URL}, which will be replaced by current plugin url (e.g. "*http://ip:port/pluginName/*"
THe example of the **manifest.json**:
```json
{
  "Label": "My Plugin",
  "Image": "{BASE_URL}logo.png",
  "URL":   "{BASE_URL}start.tengo"
}
```
### Plugins execution
Plugins are executed on client's HTTP request to the ServeMSX: **http://{ip}:{port}/{plugin}/[path]**, where **[path]** is compared to the path in the plugin's folder (plugins/**{plugin}**/**[path]**):
- if the **[path]** is omited or does not correspond to the existing file of the plugin's folder, **main.tengo** is executed;
- if the **[path]** corresponds to the directory of the plugins's folder, **main.tengo** in the directory is executed;
- if the **[path]** corresponds to the **.tengo** file, the file is executed;
- if the **[path]** correspands to any other existing file, the file is sent to the client.

For example in the structure described above, the requests:
- **http://{ip}:{port}/myplugin/** - executes main.tengo,
- **http://{ip}:{port}/myplugin/opa** - executes main.tengo,
- **http://{ip}:{port}/myplugin/main.tengo** - executes main.tengo,
- **http://{ip}:{port}/myplugin/start.tengo** - executes start.tengo,
- **http://{ip}:{port}/myplugin/logo.png** - responds woth the logo.png.
## Tengo scripting
Plugin's scripts on server side have to be written on tengo language (see the description of the language: [github.com/d5/tengo](https://github.com/d5/tengo)).
### Tengo language references
- [Language Syntax](https://github.com/d5/tengo/blob/master/docs/tutorial.md)
- [Runtime Types](https://github.com/d5/tengo/blob/master/docs/runtime-types.md) and [Operators](https://github.com/d5/tengo/blob/master/docs/operators.md)
- [Builtin Functions](https://github.com/d5/tengo/blob/master/docs/builtins.md)
- [Standard Library](https://github.com/d5/tengo/blob/master/docs/stdlib.md) (**due to security reasons, "os" module is excluded from ServeMSX**)
### Additional builtin function
One builtin function is added:

`panic({any})`, where if {any} is *undefined* it does nothing, else stops the program and if:
- {any} is *int* - returns HTTP status *int* to the client and logs it;
- {any} is other - returns HTTP status 500 with body **string({any})** to the client ang logs **string({any})**.
### Additional module "server"
#### Usage
```
srv := import("server")
```
#### Properties
- `proto {string}`: the protocol version for the request.
- `method {string}`: the HTTP method (GET, POST, PUT, etc.) of the request.
- `host {string}`: the host on which the URL is sought.
- `remote_addr {string}`: the network address that sent the request.
- `header {map of arrays of strings}`: the request header fields.
- `uri {string}`: unmodified request-target of the Request-Line (RFC 7230, Section 3.1.1) as sent by the client to a server.
- `version {string}`: the current version of ServeMSX.
- `script {string}: the name of script file is running.
- `plugin {string}`: the name of the plugin is running.
- `path {string}`: the path of the request's uri (after plugin name) (e.g. http://{ip}:{port}/{plugin}/**[path]**).
- `base_url {string}`: the base url of the plugin (e.g. **http://{ip}:{port}/{plugin}/**).
- `torrserver {string}`: the address of the TorrServer (if it is not set - empty string).

#### Functions
- `read() => {bytes/error}`: returns the request body as is.
- `read({bool}) => {map of arrays of strings}`: parses the request and:
	- if {bool} is falsy, returns the fields of the query, POST/PUT body parameters take precedence over URL query string values;
	- else returns the fields of the POST/PUT body parameters.
- `read({string}[, bool]) => {string/undefined}`: parses the request and returns the first value of the named {string} field of:
	- if [bool] is falsy or absent: the query, POST/PUT body parameters take precedence over URL query string values;
	- else: the POST/PUT body parameters.
- `write([any]...) => {undefined/error}`: writes the response, where:
	- if an argument is {map} sets the headers of the response, it should be map of arrays of strings, and should be set before status or body writings;
	- if an argument is {int} set the http status of the response to the {int} code, it should be set before body writings, if {int} between 300 & 399 and the next argument is {string} sends the response as redirect with the code = {int} and the url = {string};
	- any other arguments are writen to the body of the response.
- `request({string}[, string/bytes/map]) => {map/error}`: does the http request and returns the answer, where:
	- first argument is the url of the request;
	- if the second argument is {string} sets the http method (default is *GET*);
	- if the second argument is {bytes} sends the request body with http method *POST*;
	- if the second argument is map, sends the request with the following parameters in the map:
		- "body" {map/bytes/string}: the body of the request (if {map} it will be encoded as form);
		- "method" {string}: the http method (if "body" absent default is *GET*, else *POST*);
		- "query" {map/string}: the uri query (if {map} it will be encoded as form);
		- "header" {map}: the header fileds of the request (map of strings/array of strings);
		- "cookies" {array};
		- "user" {string};
		- "pass" {string};
		- "follow" {bool}: if it is falsy it does not follow the redirects (default is follows up to 10 times);
		- "timeout" {int}: in seconds;
	- the answer contains the response as a map of:
	 	- "status" {int}: status code;
	 	- "user" {string};
	 	- "pass" {string};
	 	- "header" {map};
	 	- "cookies" {array};
	 	- "body" {bytes};
	 	- "size" {int}: number of bytes;
	 	- "url" {string}: the final url (after all redirects).
- `encode_uri({string/map}[, bool]) => {string}`: encodes:
	- if the first parameter is map, it encodes the structure of url to string, the map should be map of arrays of string;
	- if the first parameter is string, it escapes string to query (to path, if bool is true).
- `decode_uri({string}[, bool]) => {string/error}`: unescapes query (path, if bool is true) to string.
- `parse_url([string]) => {map/error}`: parses url (if argument is absent url = url of the request) to components and returns map:
	- "scheme" {string},
	- "opaque" {string},
	- "user" {string},
	- "host" {string},
	- "path" {string}: unescaped path,
	- "raw_path" {string}: escaped path,
	- "query" {map of arrays of strings}: decoded query,
	- "raw_query" {string}: encoded query,
	- "fragment" {string}: unescaped fragment,
	- "raw_fragment" {string}: escaped fragment.
- `resolve_url({string}[, string]) => {string/error}`: resolves the url links:
	- if there are two arguments, first is the base url, second is the reference to solve;
	- if there is one argument, base url is the request url ("http://" + host + uri), and the argument is the reference to resolve.
- `memory([any]) [any]`: if [any] omited - reads from the memory, else writes to the memory. The memory keeps the object until ServeMSX is restarted or stoped.
- `file({string}[,any]) [bytes/error]`: read/write file, where:
	- if [any] omited - reads the file with the path {string} and returns {bytes/error}, if the file is not exists returns *undefined*;
	- if [any] is *undefuned* - removes the file, and returns {undefined/error};
	- else - writes {any} to the file, if it is not exists it is created, if exits - it is truncated.
- `player({string}[, bool/string]) {string/map}`: *underconstruction*.
- `log_err([any...])`: logs [any...] to the error logging.
- `log_inf([any...])`: logs [any...] to the info logging.
- `dictionary() [string]`: returns the name of current dictionary (language).
### Plugins distribution
Plugin can be distributed by the file, which can be installed by SeveMSX WEB UI. The file have to be:
- Format: gziped tar
- Name: **{plugin}.{PPP}.tgz**, where: 
	- **{plugin}** - the name of plugin (the same as plugin folder name described above);
	- **{PPP}** - the three digits of *minimum plugin engine version of ServerMSX*, it the last three digits of ServeMSX version (the version of ServeMSX format is: **M.VV.PPP**, where **M** is a digit of major version, **VV** is two digits of functionality version, **PPP** is three digits of *plugin engine version*)
- Content: the content of the plugin folder (described above), the example is:
	- :file_cabinet: **myplugin.001.tgz**:
		- :spiral_notepad: manifest.json
		- :spiral_notepad: start.tengo
		- :spiral_notepad: main.tengo
		- :framed_picture: logo.png

## HTTP API
HTTP requests http://{IP}:{PORT}/**[URI]**, where the **[URI]** can be:
|[URI]		|Request Method	|Parameters<br>(bold are required)	|Answer type	|Description|
| :---:   	| :---:	| :---:	| :---:   | :---:   |
|		|GET	| |HTML	|Web-UI of ServeMSX|
|logo.png	|GET	| |PNG	|Small logo of ServeMSX|
|logotype.png	|GET	| |PNG	|Wide logotype of ServeMSX|
|restart	|GET	| |JSON	|Restarts ServeMSX|
|settings	|GET	| |JSON	|Returns ServeMSX settings|
|msx/dictionary.json|GET| |JSON	|Returns the current dictionary of ServeMSX|
|msx/video/[path]	|GET	| |JSON	|Returns the content of video folder (My video)|
|msx/music/[path]	|GET	| |JSON	|Returns the content of music folder (My music)|
|msx/photo/[path]	|GET	| |JSON	|Returns the content of photo folder (My photo)|
|msx/torr	|GET	|link=*url/magnet/hash*<br>ttl=*title*<br>img=*url*	|JSON	|If *link* is not set, returns the content of TorrServer's torrents (My torrents).<br>If *link* is set, returns the content of the torrent, if *ttl* & *img* are set, the title and url to poster will be added if user press "Add torrent to My torrents" (yellow) button.|
|msx/input	|POST	|JSON object with properties:<br>**"action":"*string*"**,<br>"headline":"*string*",<br>"extension":"*string*",<br>"value":"*string*"	|JSON	|Returns a panel with keyboard asking user to enter a text. Afer user entered the text it executes the "action" with POST request with JSON: `{"data":"text"}`|
