# Requirements



Infrastructure

- Capture and process multiple video and audio streams
- Capable of utilising hardware acceleration for blending, compositing, scaling, and encoding of video.
- Dynamic scaling based on incoming video signals.
- Encryption of media streams
- Runs on off-the-shelf hardware
- Flexible enough to be integrated into existing AV systems. Therefore support
  for HDMI or SDI, analog audio via XLR and digital audio via Dante.
- No proprietary vendor lock in
- Affordable and Reliable
- Low Maintance


Video Portal

Generally any video portal that can process SRT and provides an iCalendar stream for scheduling.
Flexiblity. 

gocast:

Closely linked to CAMPUSOnline, the campus management system deployed at TUM.
Lightweight and scalable with a focus on livestreaming.

- Automatic lecture scheduling
- Identity Management
- Live streaming from lecture halls
- Inject self-streams with OBS (Open Broadcaster Software) or similar software for live streaming.
- Dashboard for teaching staff
    - Schedule streams
    - Access management
- Logical seperation between schools


OpenCast:
- 
OpenCast is more than a video portal. It is a general solution with integrated
support for AWS and Azure, authentication and authorization providers,
Integration for external streaming services and YouTube publication.
integration with third-party services such as Moodle. Transscription services
such as a Google Speech, IBM Watson, Azure, and Whisper. There is not one
player or one portal. Built-in video editor.

Interestly enough, the OpenCast Administration Guide completely lacks the
acquisition of audio and video. The OpenCast website links to pyCA and
commercial solution by Extron and Matrox. Unless you have unfinite money,
equiping all your lecture halls and large seminar rooms with enterprise (tm)
gear is unrealistic. pyCA lacks documentation and example hardware and
infrastructure setup.  If you want to roll your own "capture agent" fleet, be
sure to properly setup up the infrastructure.

That being said, we want to be a viable alternative to pyCA. This requires both
changes in OpenCast to support SRT streams with AV streams encapsulated in
Matroska.
