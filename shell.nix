with (import <nixpkgs> {});
mkShell {
  buildInputs = [
    # GStreamer
    gst_all_1.gstreamer
    gst_all_1.gstreamer.dev
    gst_all_1.gst-plugins-ugly # For x264enc element
    gst_all_1.gst-plugins-bad # For intervideo* elements
    gst_all_1.gst-plugins-base
    gst_all_1.gst-plugins-good
    gst_all_1.gst-libav # For avenc_aac
    gst_all_1.gst-vaapi

    libcap
    go
    gcc
    gdb

    glib
    glib.dev
    pkg-config

    # vaapi (vainfo)
    libva-utils

    alsa-utils
  ];

  shellHook = ''
    echo "Entering the captureagent development environment..."
  '';
}
