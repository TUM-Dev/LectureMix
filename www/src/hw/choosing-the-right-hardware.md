# Choosing the right Hardware

- Welcome to the world of expensive capture cards. We have seen everything from:
    - Cannot handle hot-plugging
    - Require you to know the input resolution in advance and completely
      tearing down a video pipeline when the input changes. Maybe it is just an
undercooked driver.
    - Weird dip switches for configuring capture card number. Are we in the 90s?
    - Unable to change the EDID of HDMI (Looking at you Blackmagic)
    - Closed-source driver or binary blobs
    - Video scaling in the Kernels Interrupt Handler
    - Firmware updates that require power cycles, not just reboots

Generally product descriptions are confusing, only mentioning the input
resolution, not the captured resolution.

- Most capture cards are based on FPGAs:
    - Blackmagic Decklink 4K Recorder: Xilinx Artix 7 XC7A100T
    - Magewell Eco Capture Dual HDMI M.2: Xilinx Artix 7 XC7A35T
    - Elgato Camlink 4K: CYUSB3014 and Lattice LFE5U-25F ([Reverse Engineering](https://wiki.apertus.org/index.php/Elgato_CAM_LINK_4K))

The Magewell Eco Capture was the card that cannot scale to a fixed output
resolution. The Elgato Camlink crashed multiple times during testing when the
input resolution changes. Based on recommendations from members of c3voc, the
video operation center of the Chaos Computer Club, we settled on Blackmagic
Decklinks for now.

But there is also a company called MacroSilicon Technology Co. Ltd.
specialising in ASICs for capturing audio and video..  For HDMI, there is the
MS2109 and MS2130. Both with HDMI 1.4b input and expose a USB UVC and UAC
interface.  For the MS2109 resolution is limited by the USB 2.0 interface,
therefore the newer MS2130 ASIC is recommended.


