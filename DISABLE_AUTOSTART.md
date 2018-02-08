# Disable Autostart
===================

# Windows
1. Type "Task Manager in the Windows Search Bar"

![Type "Task Manager in the Windows Search Bar"](https://raw.githubusercontent.com/arduino/arduino-create-agent/devel/images/windows/01.png)
2. Select the Startup tab

![Select the Startup tab](https://raw.githubusercontent.com/arduino/arduino-create-agent/devel/images/windows/02.png)
3. Select the autostart file

![Select the autostart file](https://raw.githubusercontent.com/arduino/arduino-create-agent/devel/images/windows/03.png)
4. Disable it

![Disable it](https://raw.githubusercontent.com/arduino/arduino-create-agent/devel/images/windows/04.png)

# Mac OSX
1. Open Finder, click on Go menu, select 'Go to Folder'

![Open Finder, click on Go menu, select 'Go to Folder'](https://raw.githubusercontent.com/arduino/arduino-create-agent/devel/images/mac/01.png)
2. Type the directory containing the autolauncher file, change <username> with your Mac username, by default the directory is /Users/username/Library/LaunchAgents

![Type the directory containing the autolauncher file, change <username> with your Mac username, by default the directory is /Users/username/Library/LaunchAgents](https://raw.githubusercontent.com/arduino/arduino-create-agent/devel/images/mac/02.png)
3. Select the ArduinoCreateAgent.plist file

![Select the ArduinoCreateAgent.plist file](https://raw.githubusercontent.com/arduino/arduino-create-agent/devel/images/mac/03.png)
4. Right click on the file name and select 'Move to Trash'

![Right click on the file name and select 'Move to Trash'](https://raw.githubusercontent.com/arduino/arduino-create-agent/devel/images/mac/04.png)

---
The command line way:
```
$ launchctl unload ~/Library/LaunchAgents/ArduinoCreateAgent.plist
```

# Linux
1. Show hidden files

![Show hidden files](https://raw.githubusercontent.com/arduino/arduino-create-agent/devel/images/linux/01.png)
2. Select the .config dir in your home

![Select the .config dir in your home](https://raw.githubusercontent.com/arduino/arduino-create-agent/devel/images/linux/02.png)
3. Select the autostart dir

![Select the autostart dir](https://raw.githubusercontent.com/arduino/arduino-create-agent/devel/images/linux/03.png)
4. Move the file to the trash

![Move the file to the trash](https://raw.githubusercontent.com/arduino/arduino-create-agent/devel/images/linux/04.png)

---
The command line way:

Just remove the autostart file in your desktop manager, in Ubuntu is:
```
$ rm $HOME/.config/autostart/arduino-create-agent.desktop
```
To start manually the agent you can open the file at:
```
$ nohup $HOME/ArduinoCreateAgent-1.1/Arduino_Create_Bridge &
```
or in the location selected during the installation