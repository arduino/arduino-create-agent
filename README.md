arduino-create-agent
====================

Arduino Create Agent is a tool that allows you to control your boards from https://create.arduino.cc

# History
arduino-create-agent is a fork of @johnlauer's serial-port-json-server (which we really want to thank for his kindness and great work)

The history has been rewritten to keep the repo small (thus removing all binaries committed in the past)

Then it was completely rewritten from scratch to address the new featureset

# Installation
Follow the getting started guide at https://create.arduino.cc/getting-started

# How does it work
Arduino Create exposes a set of rest api hosted on a variable port on localhost (eg http://localhost:8990/v1) to detect connected boards, install programming tools and program boards

It also allows to open a websocket connection to a board connected through the serial port.



## Security
A web server running on localhost makes a lot of people nervous. That's why we thought hard about security:

- You can disable autostart and run the Agent only when needed: see [Disable Autostart](DISABLE_AUTOSTART.md)
- To prevent malicious websites to perform request on the Agent through your browser, we use [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS). You can control who has access through the `origin` field on the configuration file: see [Configure](#configure)
- Every url that needs to be downloaded, or command that needs to be ran has to be signed by a trusted party. You can control who is trusted through the `trusted` folder: see [Configure](#configure)
- Commands are whitelisted, so even a trusted party can't execute a malicious command. You can control which commands are whitelisted through the `whitelist` field of the configuration file: see [Configure](#configure)

# Configure
## config.ini
Arduino Create Agent uses `.ini` files to read its options. In your install folder you'll find a configuration files that's already tuned for everyday use with create:

```ini
origins: https://create.arduino.cc
whitelist: avrdude,bossac
```

- origins is a comma-separated list of website that are allowed to contact the Agent
- whitelist is the list of commands they are allowed to perform on your machine

## trusted folder
The trusted folder contains the public keys of the trusted origins. So if you have `https://create.arduino.cc` you should have its public key on the trusted folder.