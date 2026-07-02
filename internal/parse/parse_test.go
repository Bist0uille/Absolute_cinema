package parse

import "testing"

// Cas réels relevés sur le disque D:\Films & Séries.
func TestParseMovies(t *testing.T) {
	cases := []struct {
		name  string
		title string
		year  int
		badge string
	}{
		{"(1976) Les naufragés de l'Ile de la Tortue.mkv", "Les naufragés de l'Ile de la Tortue", 1976, ""},
		{"12 Angry Men (12 hommes en colère) DVDRip VOSTFR.avi", "12 Angry Men (12 hommes en colère)", 0, "VOSTFR"},
		{"Ali.2001.DVDRip.{x264+HE-AAC-V2}{Fr-Eng}{Sub.Fr-Eng}-™.mkv", "Ali", 2001, "MULTI"},
		{"Amours chiennes (2000) (Amores perros) 720p x264 AAC 5.1 MULTI [NOEX].mkv", "Amours chiennes", 2000, "MULTI"},
		{"Better Watch Out 2017 1080p FR EN X264 AC3-mHDgz.mkv", "Better Watch Out", 2017, "MULTI"},
		{"Black.Swan.2010.DVDSCR.XviD-ViSiON.avi", "Black Swan", 2010, ""},
		{"Bound (Unrated Cut)_1996_Wachowski siblings_720p_VOST FR.mkv", "Bound", 1996, "VOSTFR"},
		{"Bruce.Tout.Puissant.2003.Multi.VFF.720P.mHD.X264.AC3-ROMKENT.mkv", "Bruce Tout Puissant", 2003, "MULTI"},
		{"Citizen Kane.avi", "Citizen Kane", 0, ""},
		{"Cleo from 5 to 7 1962 1080p FR X264 AAC-mHDgz.mkv", "Cleo from 5 to 7", 1962, "VF"},
		{"Darkest Hour (2017) VFQ-ENG AC3 BluRay 1080p x264.GHT.mkv", "Darkest Hour", 2017, "MULTI"},
		{"Eternal.sunshine.of.the.spotless.mind.2004.DVDRip.{Rv9+He-Aac}{Fr-Eng}{Sub.Fr-Eng-Spa}[XCT].mkv", "Eternal sunshine of the spotless mind", 2004, "MULTI"},
		{"Full Metal Jacket.avi", "Full Metal Jacket", 0, ""},
		{"Futurama.Into.The.Wild.Green.Yonder.2009.MULTi.1080p.BluRay.AC3.x264-STEGNER.mkv", "Futurama Into The Wild Green Yonder", 2009, "MULTI"},
		{"J.Ai.Tue.Ma.Mere.FRENCH.DVDRiP.XViD-JSK.avi", "J Ai Tue Ma Mere", 0, "VF"},
		{"L'homme Bicentenaire.avi", "L'homme Bicentenaire", 0, ""},
		{"Le Fabuleux destin d'Amélie Poulain.1080p.HEVC.NoTag.mkv", "Le Fabuleux destin d'Amélie Poulain", 0, ""},
		{"Le Garcon et la Bete 2015 1080p FR JP X264 AC3-mHDgz.mkv", "Le Garcon et la Bete", 2015, "MULTI"},
		{"Le.Septieme.Sceau.1957.VOSTFR.DVDRip.XviD.AC3-NccL.avi", "Le Septieme Sceau", 1957, "VOSTFR"},
		{"Le.Voyage.fantastique.(Fantastic.Voyage).1966.BluRay.HDLight.1080p.Multi.x264.AC3-Arcadia.mkv", "Le Voyage fantastique (Fantastic Voyage)", 1966, "MULTI"},
		{"Les Valseuses (1974)   1080p FR x264 ac3 mHDgz.mkv", "Les Valseuses", 1974, "VF"},
		{"Les.Goonies.1985.TRUEFRENCH.Blu-Ray.x265.HEVC.AC3-STARLIGHTER.mkv", "Les Goonies", 1985, "VF"},
		{"Limitless.Unrated.Extended.Cut.2011.DVDRip.{x264+HE-AAC.5.1}{Fr-Eng}{Sub.Fr-Eng}.mkv", "Limitless", 2011, "MULTI"},
		{"Manhattan (1979) 720p BluRay x254-vsenc.mkv", "Manhattan", 1979, ""},
		{"Miraï, ma petite soeur (2018).mkv", "Miraï, ma petite soeur", 2018, ""},
		{"Old.Boy.2003.VFF.720p.mHD.AC3.x264-ROMKENT.mkv", "Old Boy", 2003, "VF"},
		{"Parasite (2019) VOST.FR-ENG WEBrip 1080p x265 AAC-JiHeff.mkv", "Parasite", 2019, "VOSTFR"},
		{"Pokémon Le Film 01 - Mewtwo contre Mew (1998).mkv", "Pokémon Le Film 01 - Mewtwo contre Mew", 1998, ""},
		{"Ready.Player.One.2018.TRUEFRENCH.1080p.10bit.BluRay.x265-NSP.mkv", "Ready Player One", 2018, "VF"},
		{"Sid et Nancy 1986.avi", "Sid et Nancy", 1986, ""},
		{"Summer.Wars.MULTI.FRENCH.1080p.BRRip.x264.SubFr.mkv", "Summer Wars", 0, "MULTI"},
		{"Sympathy.For.Mr.Vengeance.2002.VOSTFR.m-720p.x264.AC3-mOe.mkv", "Sympathy For Mr Vengeance", 2002, "VOSTFR"},
		{"The Big Short 2015 MULTi VFF AC3 1080p HDLight x264.GHT (Le casse du siecle).mkv", "The Big Short", 2015, "MULTI"},
		{"Transcendance (2014) [1080p] MULTi VFF BluRay x264-PopHD.mkv", "Transcendance", 2014, "MULTI"},
		{"War Dogs (2016) [720p] VOSTFR BluRay x264 DTS 5.1-FRUITED.mkv", "War Dogs", 2016, "VOSTFR"},
		{"Zazie.Dans.Le.Metro.1960.FRENCH.BRRiP.x264.AC3-CiRAR.mkv", "Zazie Dans Le Metro", 1960, "VF"},
		{"Taken.2 2012", "Taken 2", 2012, ""},
		{"Terminator.1.1984.MULTi.1080p.BluRay.x264-PopHD.mkv", "Terminator 1", 1984, "MULTI"},
		{"Salyut 7 2017 1080p FR RU X264 AC3-mHDgz.mkv", "Salyut 7", 2017, "MULTI"},
		{"Interstellar 2014 BR EAC3 VFF VO 1080p x265 10Bits T0M.mkv", "Interstellar", 2014, "MULTI"},
		{"Dune 1984 BDRip x264 1080p AC3 5.1 MULTI (VFF-VO).mkv", "Dune", 1984, "MULTI"},
		{"Inception.2010.TrueFrench.1080p.HDLight-x264.GHT.mkv", "Inception", 2010, "VF"},
		{"Rebecca (1940, Alfred Hitchcock)", "Rebecca", 1940, ""},
		{"À.Bout.De.Souffle.1960.1080P.x265.Fr.AAC.mkv", "À Bout De Souffle", 1960, "VF"},
		{"Snowpiercer.2013.VOSTFR.BRRiP.XviD.AC3-S.V.avi", "Snowpiercer", 2013, "VOSTFR"},
		{"Requiem for a dream.2000.French.1080p.mHD.x264.AC3 5.1-touriste.mkv", "Requiem for a dream", 2000, "VF"},
		{"Phantom of the paradise (1974).BDRip.HDLight 1080p.HEVC.x265.LC-AAC 5.1 VFF & VOST-LuX.mkv", "Phantom of the paradise", 1974, "MULTI"},
		{"Batman Begins 1080p VO-VF STFr-Eng.mkv", "Batman Begins", 0, "MULTI"},
		{"The Dark Knight IMAX 1080p VO-VF STFr-Eng.mkv", "The Dark Knight", 0, "MULTI"},
		{"Apocalypse Now Redux (1979) [1080p] x264 - Jalucian.mp4", "Apocalypse Now", 1979, ""},
		{"1988 - PHOTO DE FAMILLE.avi", "PHOTO DE FAMILLE", 1988, ""},
		{"Il était une forêt (2013)", "Il était une forêt", 2013, ""},
		{"Kamikaze assaut dans le pacifique.FRENCH.avi.avi", "Kamikaze assaut dans le pacifique", 0, "VF"},
		{"Death.in.Venice.1971.MULTI.PAL.DVDR-Atreyou8.ISO", "Death in Venice", 1971, "MULTI"},
		{"3 Idiots (2009) VOSTFR h265.mkv", "3 Idiots", 2009, "VOSTFR"},
		{"Paprika.2006.BLURAY.1080p.X265.HEVC.MULTI.VFF.DTS-HD.MA.5.1-JACKT.mkv", "Paprika", 2006, "MULTI"},
		{"Albator Corsaire de l'Espace [1080p] MULTi 2013 BluRay x264-Pop.mkv", "Albator Corsaire de l'Espace", 2013, "MULTI"},
		{"On a marché sur la Lune - FR -HQ.avi", "On a marché sur la Lune", 0, "VF"},
		{"West Side Story (1961) (gixerk9).mp4", "West Side Story", 1961, ""},
	}
	for _, c := range cases {
		p := ParseName(c.name)
		if p.Title != c.title {
			t.Errorf("%q: titre = %q, attendu %q", c.name, p.Title, c.title)
		}
		if p.Year != c.year {
			t.Errorf("%q: année = %d, attendu %d", c.name, p.Year, c.year)
		}
		if got := p.Badge(); got != c.badge {
			t.Errorf("%q: badge = %q, attendu %q (tags=%v audio=%v)", c.name, got, c.badge, p.LangTags, p.AudioLangs)
		}
	}
}

func TestParseEpisodes(t *testing.T) {
	cases := []struct {
		name    string
		title   string
		season  int
		episode int
	}{
		{"Breaking.Bad.S01E01.avi", "Breaking Bad", 1, 1},
		{"Game.of.Thrones.S01E05.720p.BluRay.450MB.ShAaNiG.com.mkv", "Game of Thrones", 1, 5},
		{"Game Of Thrones 08x05 MULTI ''Les Cloches'' ... WebDl1080p ! 2019.mkv", "Game Of Thrones", 8, 5},
		{"Sherlock 1x01 A Study In Pink HDTV XviD-FoV", "Sherlock", 1, 1},
		{"Black.Mirror.S03E01.PROPER.1080p.NF.WEBRip.DD5.1.HEVC.x265.sharpysword.mkv", "Black Mirror", 3, 1},
		{"Vikings - 1x04 - Trial.x264 2HD.fr.srt", "Vikings", 1, 4},
		{"The.Man.In.The.High.Castle.S01E09.FASTSUB.VOSTFR.720p.WEBRip.HEVC.H265-Yn1D.mkv", "The Man In The High Castle", 1, 9},
		{"Narcos.S01E10.720p.WEBRip.x264-TASTETV.mkv", "Narcos", 1, 10},
		{"Final.Space.S02E03.720p.AMZN.WEBRip.x264-GalaxyTV.mkv", "Final Space", 2, 3},
	}
	for _, c := range cases {
		p := ParseName(c.name)
		if !p.HasEpisode {
			t.Errorf("%q: épisode non détecté", c.name)
			continue
		}
		if p.Title != c.title || p.Season != c.season || p.Episode != c.episode {
			t.Errorf("%q: (%q, S%d, E%d), attendu (%q, S%d, E%d)", c.name, p.Title, p.Season, p.Episode, c.title, c.season, c.episode)
		}
	}
}

func TestParseSeasonPacks(t *testing.T) {
	cases := []struct {
		name   string
		title  string
		season int
	}{
		{"Breaking Bad Saison 1", "Breaking Bad", 1},
		{"Black.Mirror.S02", "Black Mirror", 2},
		{"Game.of.Thrones.S01.720p.BluRay.x264.ShAaNiG", "Game of Thrones", 1},
		{"mr. robot saison02", "mr robot", 2},
		{"Mr.Robot - SAISON01 - VOSTFR - [1080p] - WEB-DL - x265 HEVC - [NOEX]", "Mr Robot", 1},
		{"Futurama saison 03 - Multi Dvdrip x264 - Haibane", "Futurama", 3},
		{"Pokémon Saison 01 La Ligue Indigo", "Pokémon", 1},
		{"Philip.K.Dicks.Electric.Dreams.S01", "Philip K Dicks Electric Dreams", 1},
		{"Sense8 - Saison 1 (1080p - x265)", "Sense8", 1},
		{"Futurama - Season 6", "Futurama", 6},
		{"The Expanse - Complete Season 1 S01 - 720p HDTV x264", "The Expanse", 1},
		{"the expanse saison 3", "the expanse", 3},
		{"Rick.&.Morty.S01.1080p.Multi.Bluray.x265-SN2P", "Rick & Morty", 1},
		{"Paradise PD.S01.TRUEFRENCH.720p.WEB-DL.x264-FTMVHD", "Paradise PD", 1},
		{"Westworld.S01.PROPER.VOSTFR.720p.WEBRiP.x264.AC3-GOBO2S", "Westworld", 1},
	}
	for _, c := range cases {
		p := ParseName(c.name)
		if !p.SeasonPack {
			t.Errorf("%q: pack de saison non détecté (parsed=%+v)", c.name, p)
			continue
		}
		if p.Title != c.title || p.Season != c.season {
			t.Errorf("%q: (%q, S%d), attendu (%q, S%d)", c.name, p.Title, p.Season, c.title, c.season)
		}
	}
}

func TestParts(t *testing.T) {
	if p := ParseName("Entre Les Murs CD1.divx"); p.Part != 1 {
		t.Errorf("CD1 non détecté: %+v", p)
	}
	if p := ParseName("Fanny och Alexander CD02.avi"); p.Part != 2 {
		t.Errorf("CD02 non détecté: %+v", p)
	}
	if p := ParseName("Godfather part II.avi"); p.Part != 0 {
		t.Errorf("'part II' ne doit PAS être une partie CD (film distinct): %+v", p)
	}
	if p := ParseName("Carlos-3-DVD-Rip-stabdou-Fr.avi"); p.Part != 0 {
		t.Errorf("'3-DVD-Rip' ne doit pas être une partie: %+v", p)
	}
}

func TestSubLang(t *testing.T) {
	cases := map[string]string{
		"Apocalypse Now Redux (1979) [1080p] x264 - Jalucian fre.srt": "fr",
		"Apocalypse Now Redux (1979) [1080p] x264 - Jalucian.srt":     "",
		"Breaking Bad - 1x01 - Pilot.DSR.0TV.en.srt":                  "en",
		"Vikings - 1x04 - Trial.x264 2HD.fr.srt":                      "fr",
	}
	for name, want := range cases {
		if got := SubLang(name); got != want {
			t.Errorf("SubLang(%q) = %q, attendu %q", name, got, want)
		}
	}
}

func TestSeasonFromFolder(t *testing.T) {
	cases := map[string]int{
		"Breaking Bad Saison 1":            1,
		"Futurama - Season 6":              6,
		"saison02":                         2,
		"Saison 01 - La ligue indigo":      1,
		"Black.Mirror.S05.COMPLETE.720p":   5,
		"Hannibal - Season 2":              2,
		"saison 1":                         1,
	}
	for name, want := range cases {
		if got := SeasonFromFolder(name); got != want {
			t.Errorf("SeasonFromFolder(%q) = %d, attendu %d", name, got, want)
		}
	}
}

func TestSimilarity(t *testing.T) {
	if s := Similarity("Les naufragés de l'Ile de la Tortue", "Les Naufragés de l'île de la Tortue"); s < 0.95 {
		t.Errorf("similarité accents = %f", s)
	}
	if s := Similarity("The Godfather", "Godfather"); s < 0.9 {
		t.Errorf("article initial = %f", s)
	}
	if s := Similarity("Taken 2", "Taken"); s > 0.9 {
		t.Errorf("Taken 2 vs Taken trop proche = %f", s)
	}
}
