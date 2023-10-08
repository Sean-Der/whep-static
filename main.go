package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/pion/webrtc/v4/pkg/media/h264reader"
)

const (
	videoFileName     = "output.h264"
	h264FrameDuration = time.Millisecond * 33
)

var (
	videoTrack *webrtc.TrackLocalStaticSample
)

func doSignaling(w http.ResponseWriter, r *http.Request) {
	offer, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		panic(err)
	}

	rtpSender, err := peerConnection.AddTrack(videoTrack)
	if err != nil {
		panic(err)
	}

	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("ICE Connection State has changed: %s\n", connectionState.String())

		if connectionState == webrtc.ICEConnectionStateFailed {
			peerConnection.Close()
		}
	})

	if err = peerConnection.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: string(offer)}); err != nil {
		panic(err)
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	} else if err = peerConnection.SetLocalDescription(answer); err != nil {
		panic(err)
	}

	<-gatherComplete

	w.Header().Add("Location", "/doSignaling")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, peerConnection.LocalDescription().SDP)
}

func sendVideo() error {
	file, err := os.Open(videoFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	h264, err := h264reader.NewReader(file)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(h264FrameDuration)
	for ; true; <-ticker.C {
		nal, err := h264.NextNAL()
		if err == io.EOF {
			ticker.Stop()
			return nil
		}
		if err != nil {
			return err
		}

		if err = videoTrack.WriteSample(media.Sample{Data: nal.Data, Duration: h264FrameDuration}); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	_, err := os.Stat(videoFileName)
	if os.IsNotExist(err) {
		panic("output.h264 was not found")
	}

	videoTrack, err = webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion")
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			if sendVideoErr := sendVideo(); sendVideoErr != nil {
				panic(sendVideoErr)
			}
		}
	}()

	http.Handle("/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/doSignaling", doSignaling)

	fmt.Println("Open http://localhost:8080 to access this demo")
	// nolint: gosec
	panic(http.ListenAndServe(":8080", nil))
}
