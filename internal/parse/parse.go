// Package parse extrait titre, année, tags de langue, résolution, édition et
// numéros de saison/épisode depuis les noms de fichiers et de dossiers.
package parse

import (
	"regexp"
	"strconv"
	"strings"
)

// Parsed est le résultat de l'analyse d'un nom de fichier ou de dossier.
type Parsed struct {
	Title      string   // titre nettoyé, prêt pour une recherche TMDB
	Year       int      // 0 si absent
	Resolution string   // "2160p", "1080p", "720p", "480p", "" si absent
	LangTags   []string // tags bruts normalisés trouvés : VOSTFR, MULTI, VFF, FR, EN…
	AudioLangs []string // codes langue déduits des tokens post-titre : fr, en, jp…
	SubLangs   []string // langues de sous-titres déduites (STFR, SUBFRENCH…)
	Edition    string   // "Redux", "Extended", "Director's Cut", "Unrated", "IMAX", "Remastered"
	Season     int      // 0 si absent
	Episode    int      // 0 si absent
	HasEpisode bool     // motif SxxEyy ou 1x01 trouvé
	SeasonPack bool     // Sxx ou "Saison N" sans numéro d'épisode
	Part       int      // numéro de CD/partie (CD1, CD2…), 0 sinon
	IsISO      bool
}

// Badge calcule le badge de langue affiché : MULTI, VOSTFR, VF, VO ou "".
func (p Parsed) Badge() string {
	tags := map[string]bool{}
	for _, t := range p.LangTags {
		tags[t] = true
	}
	frAudio := tags["VF"] || tags["VFF"] || tags["VFQ"] || tags["VFI"] || tags["FRENCH"] || tags["TRUEFRENCH"] || tags["FR"] || contains(p.AudioLangs, "fr")
	foreign := tags["EN"] || tags["JP"] || tags["IT"] || tags["ES"] || tags["VO"] || tags["VOSTA"] || tags["ENG-SPA"]
	for _, l := range p.AudioLangs {
		if l != "fr" {
			foreign = true
		}
	}
	frSubs := tags["VOSTFR"] || tags["VOST"] || tags["STFR"] || tags["SUBFRENCH"] || tags["SUBFORCED"] || tags["FASTSUB"] || contains(p.SubLangs, "fr")

	switch {
	case tags["MULTI"] || tags["VO-VF"] || (frAudio && (foreign || tags["VOSTFR"] || tags["VOST"])):
		return "MULTI"
	case tags["VOSTFR"] || tags["VOST"] || (frSubs && !frAudio):
		return "VOSTFR"
	case frAudio:
		return "VF"
	case foreign:
		return "VO"
	}
	return ""
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

var videoExts = map[string]bool{
	".mkv": true, ".avi": true, ".mp4": true, ".m4v": true, ".mov": true,
	".wmv": true, ".mpg": true, ".mpeg": true, ".divx": true, ".ts": true,
	".webm": true, ".iso": true, ".flv": true, ".vob": true,
}

var subExts = map[string]bool{
	".srt": true, ".sub": true, ".ass": true, ".ssa": true, ".vtt": true, ".idx": true,
}

// IsVideoExt indique si l'extension (avec point, insensible à la casse) est une vidéo.
func IsVideoExt(ext string) bool { return videoExts[strings.ToLower(ext)] }

// IsSubExt indique si l'extension est un sous-titre.
func IsSubExt(ext string) bool { return subExts[strings.ToLower(ext)] }

var (
	reWebsite   = regexp.MustCompile(`(?i)\[?\s*www\.[a-z0-9.-]+\.[a-z]{2,}\s*\]?|\((?:elitetorrent|glowgaze)[^)]*\)|\[\s*(?:nextorrent|cpasbien|omgtorrent|yts\.[a-z]+|torrent9|rarbg)[^\]]*\]`)
	reEpisode   = regexp.MustCompile(`(?i)\bS(\d{1,2})[\s._-]*E(\d{1,3})(?:[-E]\d{1,3})?\b`)
	reEpisodeX  = regexp.MustCompile(`(?i)\b(\d{1,2})x(\d{2,3})\b`)
	reSeasonTag = regexp.MustCompile(`(?i)\b(?:saison|season)[\s._-]*(\d{1,2})\b|\bS(\d{1,2})\b`)
	reYearParen = regexp.MustCompile(`[(\[](19[0-9]\d|20[0-4]\d)(?:[^)\]]*)[)\]]`)
	reYearBare  = regexp.MustCompile(`\b(19[0-9]\d|20[0-4]\d)\b`)
	reCDPart    = regexp.MustCompile(`(?i)\b(?:cd|dvd|disc|disk)[\s._-]?(\d{1,2})\b`)
	reBrackets  = regexp.MustCompile(`\[[^\]]*\]|\{[^}]*\}`)
	reSpaces    = regexp.MustCompile(`\s+`)
	reDotJoin   = regexp.MustCompile(`([\pL\pN])\.([\pL\pN])`)
	reTrailerGrp = regexp.MustCompile(`-[A-Za-z0-9™]{2,20}$`)
)

// Tags techniques : leur première occurrence marque la fin du titre.
var reTechTag = regexp.MustCompile(`(?i)\b(` +
	`(?:2160|1080|720|576|480)p?|4k|uhd|3d|hsbs|10bits?|8bits?|` +
	`blu-?ray|b[dr]rip|brrip|web-?dl|webrip|hdtv(?:rip)?|dvdrip|dvdriip|dvdscr|dvdr|dvd|hdlight|m-?hd|remux|dsr|hdrip|hd-?ts|hdcam|camrip|screener|pal|ntsc|` +
	`x\.?26[45]|h\.?26[45]|hevc|xvid|x264|x265|divx|av1|avc|rv9|` +
	`ac-?3|e-?ac-?3|eac3|aac(?:[.-]?v?2)?|he-aac|lc-aac|dts(?:-?hd)?(?:[. ]?ma)?|dd\+?[257]\.?[01]|ddp[257]\.?[01]|[257]\.[01]|6ch|2ch|mp3|flac|` +
	`vostfr|vosta|vost|multi|vff|vfq|vfi|truefrench|french|subfrench|subforced|fastsub|stfr|sttfr|vf|vo|` +
	`extended|unrated|remastered|redux|imax|proper|repack|limited|internal|complete|custom|criterion|uncut|hc|` +
	`dvdrip|bdrip|webdl|jap(?:anese)?|italian|fr|` +
	`director'?s[\s._-]?cut|dc` +
	`)\b`)

// Tags de langue reconnus, normalisés en majuscules.
var reLangTag = regexp.MustCompile(`(?i)\b(vostfr|vosta|vost[\s._]?fr|vost|multi|vff|vfq|vfi|truefrench|french|subfrench|subforced|fastsub|stt?[\s._]?fr(?:-eng)?|vf|vo-?vf|vo|fr-?eng?|eng?-?fr|eng-spa|jap-fr|fr|eng?|jp|japanese|italian|spa(?:nish)?)\b`)

// Éditions à conserver dans les métadonnées.
var reEdition = regexp.MustCompile(`(?i)\b(redux|extended|unrated|remastered|imax|director'?s[\s._-]?cut|criterion|version[\s._-]?longue)\b`)

var reResolution = regexp.MustCompile(`(?i)\b(2160|1080|720|576|480)p?\b|\b(4k)\b`)

// codes langue bruts trouvés en zone technique -> ISO 639-1
var langCodeMap = map[string]string{
	"fr": "fr", "fre": "fr", "fra": "fr", "french": "fr", "truefrench": "fr", "vff": "fr", "vfq": "fr", "vfi": "fr", "vf": "fr",
	"en": "en", "eng": "en", "english": "en",
	"jp": "ja", "jap": "ja", "japanese": "ja",
	"es": "es", "spa": "es", "spanish": "es",
	"it": "it", "italian": "it", "ita": "it",
	"ru": "ru", "de": "de", "ger": "de", "swe": "sv", "ko": "ko", "kor": "ko", "zh": "zh",
}

// ParseName analyse un nom de fichier (avec ou sans extension) ou de dossier.
func ParseName(name string) Parsed {
	var p Parsed

	// extension
	if i := strings.LastIndex(name, "."); i > 0 {
		ext := strings.ToLower(name[i:])
		if videoExts[ext] {
			p.IsISO = ext == ".iso"
			name = name[:i]
			// double extension (".avi.avi")
			if j := strings.LastIndex(name, "."); j > 0 && videoExts[strings.ToLower(name[j:])] {
				name = name[:j]
			}
		}
	}

	name = reWebsite.ReplaceAllString(name, " ")

	// saison / épisode avant toute normalisation
	if m := reEpisode.FindStringSubmatchIndex(name); m != nil {
		p.Season, _ = strconv.Atoi(name[m[2]:m[3]])
		p.Episode, _ = strconv.Atoi(name[m[4]:m[5]])
		p.HasEpisode = true
	} else if m := reEpisodeX.FindStringSubmatchIndex(name); m != nil {
		p.Season, _ = strconv.Atoi(name[m[2]:m[3]])
		p.Episode, _ = strconv.Atoi(name[m[4]:m[5]])
		p.HasEpisode = true
	}

	// CD / partie
	if m := reCDPart.FindStringSubmatch(name); m != nil {
		// "3-DVD-Rip" n'est pas une partie : exiger que le tag soit collé à un nombre <= 8
		if n, _ := strconv.Atoi(m[1]); n >= 1 && n <= 8 {
			if !regexp.MustCompile(`(?i)\bdvd[\s._-]?` + m[1] + `?\s*rip`).MatchString(name) {
				p.Part = n
			}
		}
	}

	// année : parenthèses prioritaires
	yearPos := -1
	if m := reYearParen.FindStringSubmatchIndex(name); m != nil {
		p.Year, _ = strconv.Atoi(name[m[2]:m[3]])
		yearPos = m[0]
	}

	// normalisation des séparateurs
	work := strings.ReplaceAll(name, "_", " ")
	if strings.Count(work, ".") >= 1 {
		work = reDotJoin.ReplaceAllString(work, "$1 $2")
		work = reDotJoin.ReplaceAllString(work, "$1 $2") // les chevauchements ("a.b.c")
	}
	work = strings.ReplaceAll(work, "·", " ")

	// tags de langue (sur tout le nom)
	seenLang := map[string]bool{}
	for _, m := range reLangTag.FindAllString(work, -1) {
		tag := normalizeLangTag(m)
		if tag != "" && !seenLang[tag] {
			seenLang[tag] = true
			p.LangTags = append(p.LangTags, tag)
		}
	}

	// édition
	if m := reEdition.FindString(work); m != "" {
		p.Edition = normalizeEdition(m)
	}

	// résolution
	if m := reResolution.FindString(work); m != "" {
		r := strings.ToLower(m)
		if r == "4k" || r == "2160" || r == "2160p" {
			p.Resolution = "2160p"
		} else {
			p.Resolution = strings.TrimSuffix(r, "p") + "p"
		}
	}

	// année en position libre si pas trouvée en parenthèses :
	// dernière occurrence plausible (pas un nombre de résolution)
	if p.Year == 0 {
		for _, m := range reYearBare.FindAllStringSubmatchIndex(work, -1) {
			y, _ := strconv.Atoi(work[m[2]:m[3]])
			// ignorer si suivi de "p" (1080p mal découpé) — reYearBare a \b donc ok
			p.Year = y
			yearPos = -2 // position recalculée plus bas sur work
		}
	}

	// point de coupe du titre : premier tag technique, épisode, ou année
	cut := len(work)
	if m := reTechTag.FindStringIndex(work); m != nil && m[0] < cut {
		cut = m[0]
	}
	if m := reEpisode.FindStringIndex(work); m != nil && m[0] < cut {
		cut = m[0]
	}
	if m := reEpisodeX.FindStringIndex(work); m != nil && m[0] < cut {
		cut = m[0]
	}
	if p.Year != 0 {
		// retrouver la position de l'année dans work
		pat := regexp.MustCompile(`[(\[]?` + strconv.Itoa(p.Year) + `\b`)
		if m := pat.FindStringIndex(work); m != nil {
			if m[0] == 0 {
				// année en tête ("(1976) Les naufragés…", "1988 - PHOTO DE FAMILLE")
				rest := work[m[1]:]
				rest = strings.TrimLeft(rest, ") ].-–—")
				if strings.TrimSpace(rest) != "" {
					work = rest
					// recalcul du cut sur le nouveau work
					cut = len(work)
					if mm := reTechTag.FindStringIndex(work); mm != nil && mm[0] < cut {
						cut = mm[0]
					}
				}
			} else if m[0] < cut {
				cut = m[0]
			}
		}
	}
	_ = yearPos

	title := work[:cut]

	// nettoyage final du titre
	title = reBrackets.ReplaceAllString(title, " ")
	title = reCDPart.ReplaceAllString(title, " ")
	if p.HasEpisode {
		title = reEpisode.ReplaceAllString(title, " ")
		title = reEpisodeX.ReplaceAllString(title, " ")
	}
	title = strings.ReplaceAll(title, "™", " ")
	title = strings.ReplaceAll(title, ".", " ")
	title = strings.ReplaceAll(title, "`", "'")
	title = stripCombining(title)
	title = strings.Trim(title, " .-–—([{,;!")
	// groupe de release résiduel en fin ("…-mHDgz") uniquement si tags techniques présents plus loin
	if cut < len(work) {
		title = reTrailerGrp.ReplaceAllString(title, " ")
	}
	title = reSpaces.ReplaceAllString(title, " ")
	title = strings.TrimSpace(title)

	// saison seule ("Saison 2", "S01", "Breaking Bad Saison 1") si pas d'épisode
	if !p.HasEpisode {
		if m := reSeasonTag.FindStringSubmatch(work); m != nil {
			num := m[1]
			if num == "" {
				num = m[2]
			}
			if n, _ := strconv.Atoi(num); n >= 1 && n <= 40 {
				p.Season = n
				p.SeasonPack = true
			}
		}
	}
	if p.SeasonPack || p.HasEpisode {
		// couper le titre avant le marqueur de saison
		if m := regexp.MustCompile(`(?i)[\s._-]*(?:-\s*)?(?:saison|season|s\d{1,2}\b)`).FindStringIndex(title); m != nil && m[0] > 0 {
			title = strings.TrimSpace(strings.Trim(title[:m[0]], " .-–—"))
		}
	}

	// langues audio déduites des tokens en zone technique ("1080p FR JP X264")
	if cut < len(work) {
		tech := reLangTag.ReplaceAllString(work[cut:], " ")
		for _, tok := range strings.Fields(tech) {
			tok = strings.Trim(strings.ToLower(tok), "()[]{}.,-&!")
			if iso, ok := langCodeMap[tok]; ok && !contains(p.AudioLangs, iso) {
				p.AudioLangs = append(p.AudioLangs, iso)
			}
		}
	}
	// combos "Fr-Eng", "VO-VF", "Jap-Fr" comptent comme audio multiple
	for _, t := range p.LangTags {
		switch t {
		case "FR-EN":
			addLang(&p.AudioLangs, "fr")
			addLang(&p.AudioLangs, "en")
		case "JAP-FR":
			addLang(&p.AudioLangs, "fr")
			addLang(&p.AudioLangs, "ja")
		case "ENG-SPA":
			addLang(&p.AudioLangs, "en")
			addLang(&p.AudioLangs, "es")
		case "STFR", "SUBFRENCH", "SUBFORCED", "VOSTFR", "FASTSUB":
			addLang(&p.SubLangs, "fr")
		}
	}

	p.Title = title
	return p
}

// stripCombining retire les diacritiques combinants (noms de fichiers en
// Unicode NFD, typiquement créés sur Mac) : "Guépard" → "Guepard".
// TMDB est insensible aux accents, la recherche reste correcte.
func stripCombining(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= 0x0300 && r <= 0x036F {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func addLang(ss *[]string, l string) {
	if !contains(*ss, l) {
		*ss = append(*ss, l)
	}
}

func normalizeLangTag(s string) string {
	t := strings.ToUpper(strings.TrimSpace(s))
	t = strings.NewReplacer(".", "", "_", "", " ", "", "-", "").Replace(t)
	switch {
	case strings.HasPrefix(t, "VOSTFR"):
		return "VOSTFR"
	case t == "VOSTA":
		return "VOSTA"
	case t == "VOST":
		return "VOST"
	case t == "MULTI":
		return "MULTI"
	case t == "VFF" || t == "VFQ" || t == "VFI":
		return t
	case t == "TRUEFRENCH" || t == "FRENCH":
		return "FRENCH"
	case t == "SUBFRENCH":
		return "SUBFRENCH"
	case t == "SUBFORCED":
		return "SUBFORCED"
	case t == "FASTSUB":
		return "FASTSUB"
	case strings.HasPrefix(t, "STFR") || strings.HasPrefix(t, "STTFR"):
		return "STFR"
	case t == "VOVF" || t == "VO-VF":
		return "VO-VF"
	case t == "FRENG" || t == "FREN" || t == "ENGFR" || t == "ENFR":
		return "FR-EN"
	case t == "ENGSPA":
		return "ENG-SPA"
	case t == "JAPFR":
		return "JAP-FR"
	case t == "VF":
		return "VF"
	case t == "VO":
		return "VO"
	case t == "FR":
		return "FR"
	case t == "EN" || t == "ENG":
		return "EN"
	case t == "JP" || t == "JAPANESE" || t == "JAP":
		return "JP"
	case t == "ITALIAN":
		return "IT"
	case t == "SPA" || t == "SPANISH":
		return "ES"
	}
	return ""
}

func normalizeEdition(s string) string {
	t := strings.ToLower(s)
	switch {
	case strings.Contains(t, "redux"):
		return "Redux"
	case strings.Contains(t, "extended"):
		return "Extended"
	case strings.Contains(t, "unrated"):
		return "Unrated"
	case strings.Contains(t, "remaster"):
		return "Remastered"
	case strings.Contains(t, "imax"):
		return "IMAX"
	case strings.Contains(t, "director"):
		return "Director's Cut"
	case strings.Contains(t, "criterion"):
		return "Criterion"
	case strings.Contains(t, "longue"):
		return "Version longue"
	}
	return ""
}

// SeasonFromFolder extrait un numéro de saison d'un nom de dossier
// ("Saison 01 - La ligue indigo", "Season 6", "saison02", "S01", "Complete Season 1 S01").
func SeasonFromFolder(name string) int {
	if m := regexp.MustCompile(`(?i)\b(?:saison|season)[\s._-]*(\d{1,2})\b`).FindStringSubmatch(name); m != nil {
		n, _ := strconv.Atoi(m[1])
		return n
	}
	if m := regexp.MustCompile(`(?i)\bS(\d{1,2})\b`).FindStringSubmatch(name); m != nil {
		n, _ := strconv.Atoi(m[1])
		return n
	}
	return 0
}

// SubLang devine la langue d'un fichier de sous-titres depuis son nom
// ("....en.srt", "... fre.srt", "..._fr.srt", "....VOSTFR.srt").
func SubLang(name string) string {
	base := name
	if i := strings.LastIndex(base, "."); i > 0 {
		base = base[:i]
	}
	m := regexp.MustCompile(`(?i)[.\s_(-](fr|fre|fra|french|vostfr|en|eng|english|es|spa|forced)[)\s]*$`).FindStringSubmatch(base)
	if m == nil {
		return ""
	}
	switch strings.ToLower(m[1]) {
	case "fr", "fre", "fra", "french", "vostfr", "forced":
		return "fr"
	case "en", "eng", "english":
		return "en"
	case "es", "spa":
		return "es"
	}
	return ""
}

var foldReplacer = strings.NewReplacer(
	"à", "a", "â", "a", "ä", "a", "á", "a", "ã", "a", "å", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"î", "i", "ï", "i", "í", "i", "ì", "i",
	"ô", "o", "ö", "o", "ó", "o", "ò", "o", "õ", "o", "ø", "o",
	"û", "u", "ü", "u", "ú", "u", "ù", "u",
	"ç", "c", "ñ", "n", "œ", "oe", "æ", "ae", "ß", "ss",
	"'", " ", "’", " ", "`", " ", ":", " ", ",", " ", "!", " ", "?", " ",
	"&", "and",
)

var reFoldSpaces = regexp.MustCompile(`\s+`)

// Fold normalise un titre pour comparaison : minuscules, accents pliés,
// articles initiaux retirés, ponctuation neutralisée.
func Fold(s string) string {
	s = strings.ToLower(s)
	s = foldReplacer.Replace(s)
	s = reFoldSpaces.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	for _, art := range []string{"the ", "le ", "la ", "les ", "l ", "un ", "une ", "der ", "el ", "il ", "a "} {
		if strings.HasPrefix(s, art) {
			s = s[len(art):]
			break
		}
	}
	return strings.TrimSpace(s)
}

// Similarity retourne une similarité [0,1] entre deux titres pliés (Levenshtein normalisée).
func Similarity(a, b string) float64 {
	a, b = Fold(a), Fold(b)
	if a == b {
		return 1
	}
	if a == "" || b == "" {
		return 0
	}
	d := levenshtein(a, b)
	max := len(a)
	if len(b) > max {
		max = len(b)
	}
	return 1 - float64(d)/float64(max)
}

func levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	prev := make([]int, len(rb)+1)
	cur := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		cur[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			cur[j] = min3(cur[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[len(rb)]
}

func min3(a, b, c int) int {
	if b < a {
		a = b
	}
	if c < a {
		a = c
	}
	return a
}
