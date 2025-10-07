# Introduction

The LectureMix project is a collection of open source software and hardware for
recording and streaming lectures. This includes the acquisition of audio and
video, processing, and distribution. LectureMix is developed by the [Open
Source @ TUM e.V](https://www.tum.dev/) and deployed at the Technical
University of Munich (TUM). The goal of the project is to provide secure and
reliable software, able to run on off-the-shelf hardware.

It started of as a small video portal from the Technical University of Munich
(TUM), primarly designed to mirror lectures across multiple lecture halls.
During this time, teaching staff interested in recording a lecture were mostly
on their own.

Mostly, because lecture halls at TUM are generally equipped with a USB
interface, exposing camera and audio as a USB webcam. This solves the problem
of acquiring a high-quality capture of the lecture, but not the recording,
streaming, or distribution of it.

This may not be much of a problem for simple video conferencing, but correctly
configuring and recording a lecture often times leads to frustration. Not only
by teaching staff wasting time setting up a recording before each lecture, but
also students being faced with corrupt or incomplete recordings.

With LectureMix, capturing of the signal happens behind the scenes. For the
teaching staff, this means that they can connect their laptop or tablet to the
HDMI or USB-C cable at the lectern as usual. A streaming server captures video
and audio channels in the background and processes them.

The encoded streams are then sent over the network to the backend of the video
portal for further distribution. Streams and recordings are scheduled based on
the timetable supplied by the campus management system. A video-on-demand (VoD)
is available shortly after the lecture ends.

As of the 28th of May 2025, the LectureMix infrastructure at TUM serviced
3,078,062 Video-on-Demand (VoD) views and 411,221 live views. It has a user
reach of 38,527 students, 845 lecturers and course administrators, with 913
total lectures. Data was captured from June 28th, 2021 onwards. You can find
the instance at https://tum.live.

This e-book caters to people interested in learning how to set up such a
service. We will introduce the project by looking at the problem first, then
introducing the software and hardware. Notice that we left out the hardest part,
namely, the initial deployment and operation of the service.

We also welcome further research, not only on the engineerng side. We as
engineers are often concered about the implementation and the operational
aspect, disregarding the social implications. How does live streaming of
lectures and/or recordings impact the students and teaching staff?


Draft Section:


- Welcoming research on TUM.live
    - Accessbility: Impairments, Changing Speed of Recordings, Stopping to make notes
    - How TUM.live affects student learning behaviour
    - Lecture Attendance and Fear of empty lecture halls. A natural indicator
      of the quality of a lecture is student attendance of streamed and
recorded lectures. How does the live streaming of lectures impact the
attendance rate?
    - Usability of the system (lecturers and students)
        - Ease of Use
        - Bandwidth required
        - Reliability
    - Expanding to a system that integrates content: Think MIT Open Course Ware
    - Capturing lectures that make use of a blackboard
    - Using live stream to see previous slides

Research:
- https://pmc.ncbi.nlm.nih.gov/articles/PMC9900091/pdf/FEB4-13-217.pdf
- https://journals.sagepub.com/doi/10.1177/21582440241305325
