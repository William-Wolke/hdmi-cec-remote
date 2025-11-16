# HTPC inputs using a tv remote

This project aims to make a "dumb" tv remote act as a mouse and keyboard for basic HTPC
needs such as running a browser in kiosk mode.

## cec-client
Using cec-client to read button presses and using ydotool to move the mouse and press keys.
Since it's just a go program you could open different apps via keypresses


## flirc v2
My panasonic tv remote has some keys reserved for a panasonic dvd box and those keypresses aren't picked up by the tv and not forwarded via cec.
To add them as extra keys I use a flirc v2 to get extra keys, however since I'm using the normal flirc software the functionality is limited to keyboard inputs. Key combinations, mouse clicks and mouse movements are restricted to the cec inputs.

## Installation

```bash
make
sudo make install
make clean
```

