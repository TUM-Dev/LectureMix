{ lib
, stdenv
, fetchFromGitHub
, buildGoModule
, pkg-config
, gcc
, makeWrapper
, gst_all_1
, glib
, blackmagic-desktop-video
}:

let
  version = "0.0.1";
in
buildGoModule {
  pname = "streamd";
  inherit version;

  # src = fetchFromGitHub {
  #   owner = "TUM-Dev";
  #   repo = "LectureMix";
  #   rev = "v${version}";
  #   hash = lib.fakeHash;
  # };
  src = ./streamd;
    
  # GStreamer loads the decklink library via dlopen. We thus need to patch the ld library path
  postInstall = ''
    wrapProgram $out/bin/streamd --prefix LD_LIBRARY_PATH : ${blackmagic-desktop-video}/lib
  '';

  nativeBuildInputs = [ pkg-config gcc makeWrapper ];
  buildInputs = [
    gst_all_1.gstreamer
    gst_all_1.gst-plugins-ugly
    gst_all_1.gst-plugins-bad 
    gst_all_1.gst-plugins-base
    gst_all_1.gst-plugins-good
    gst_all_1.gst-vaapi
    glib
    blackmagic-desktop-video
  ];

  #vendorHash = lib.fakeHash;
  vendorHash = "sha256-4NkjfWhA4m3x4byOiKAx1NberaaYETAjX7DJDcguiVQ=";


  meta = with lib; {
    homepage = "https://github.com/TUM-Dev/streamd";
    description = "An open-source live-streaming stack for recording and streaming lectures";
    license = licenses.gpl2Only;
    maintainers = with maintainers; [ hmelder ];
  };
}