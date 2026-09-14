package ascii
const (
	BannerShadow="shadow"
	BannerStandard="standard"
	BannerThinkertoy="thinkertoy"
)
type Generator struct{
	mu    sync.RWMutex
	dir   string 
	fonts map[string]*font
}
type font struct{
	height int 
	glyphs map[rune][]string
}
var ValidBanners = []string{BannerStandard, BannerShadow, BannerThinkertoy}	
//to say we are at the current working directory
func Newgenerator(dir string) *Generator {
	return &Generator{
		dir: dir,
		fonts: make(map[string]*font),
	}
}