package main

var builtinStations = []*RadioStation{
	{Name: "Groove Salad", URL: "https://ice1.somafm.com/groovesalad-256-mp3", Genre: "Ambient", Country: "US", Bitrate: "256k"},
	{Name: "Secret Agent", URL: "https://ice1.somafm.com/secretagent-128-mp3", Genre: "Lounge", Country: "US", Bitrate: "128k"},
	{Name: "Drone Zone", URL: "https://ice1.somafm.com/dronezone-256-mp3", Genre: "Ambient", Country: "US", Bitrate: "256k"},
	{Name: "Lush", URL: "https://ice1.somafm.com/lush-128-mp3", Genre: "Vocal Lounge", Country: "US", Bitrate: "128k"},
	{Name: "Indie Pop Rocks", URL: "https://ice1.somafm.com/indiepop-128-mp3", Genre: "Indie Pop", Country: "US", Bitrate: "128k"},
	{Name: "Metal Detector", URL: "https://ice1.somafm.com/metal-128-mp3", Genre: "Metal", Country: "US", Bitrate: "128k"},
	{Name: "Deep Space One", URL: "https://ice1.somafm.com/deepspaceone-256-mp3", Genre: "Ambient", Country: "US", Bitrate: "256k"},
	{Name: "Folk Forward", URL: "https://ice1.somafm.com/folkfwd-128-mp3", Genre: "Folk", Country: "US", Bitrate: "128k"},
	{Name: "Reggae Cafe", URL: "https://ice1.somafm.com/reggae-128-mp3", Genre: "Reggae", Country: "US", Bitrate: "128k"},
	{Name: "Seven Inch Soul", URL: "https://ice1.somafm.com/7soul-128-mp3", Genre: "Soul/R&B", Country: "US", Bitrate: "128k"},
	{Name: "Beat Blender", URL: "https://ice1.somafm.com/beatblender-128-mp3", Genre: "Electronic", Country: "US", Bitrate: "128k"},
	{Name: "Underground 80s", URL: "https://ice1.somafm.com/u80s-256-mp3", Genre: "80s", Country: "US", Bitrate: "256k"},
	{Name: "BBC Radio 1", URL: "https://stream.live.vc.bbcmedia.co.uk/bbc_radio_one", Genre: "Pop/Dance", Country: "UK", Bitrate: "128k"},
	{Name: "BBC Radio 2", URL: "https://stream.live.vc.bbcmedia.co.uk/bbc_radio_two", Genre: "Mixed", Country: "UK", Bitrate: "128k"},
	{Name: "BBC Radio 4", URL: "https://stream.live.vc.bbcmedia.co.uk/bbc_radio_fourfm", Genre: "Talk/Drama", Country: "UK", Bitrate: "128k"},
	{Name: "BBC World Service", URL: "https://stream.live.vc.bbcmedia.co.uk/bbc_world_service", Genre: "News/Talk", Country: "UK", Bitrate: "96k"},
	{Name: "KEXP 90.3", URL: "https://kexp-mp3-128.streamguys1.com/kexp128.mp3", Genre: "Indie/Alt", Country: "US", Bitrate: "128k"},
	{Name: "Nashe Radio", URL: "http://nashe1.hostingradio.ru:80/nashe-128.mp3", Genre: "Rock", Country: "RU", Bitrate: "128k"},
}
