package pbuff

import "github.com/vespian/go-exercises/xkcd/pkg/types"

func PBStoryFromStory(s *types.Story) *PBStory {
	res := PBStory{
		Alt:        s.Alt,
		Day:        int32(s.Day),
		Img:        s.Img,
		Link:       s.Link,
		Month:      int32(s.Month),
		News:       s.News,
		Num:        int64(s.Num),
		SafeTitle:  s.SafeTitle,
		Title:      s.Title,
		Transcript: s.Transcript,
		Year:       int32(s.Year),
	}

	return &res
}

func StoryFromPBStory(p *PBStory) *types.Story {
	res := types.Story{
		Alt:        p.Alt,
		Day:        int(p.Day),
		Img:        p.Img,
		Link:       p.Link,
		Month:      int(p.Month),
		News:       p.News,
		Num:        int(p.Num),
		SafeTitle:  p.SafeTitle,
		Title:      p.Title,
		Transcript: p.Transcript,
		Year:       int(p.Year),
	}

	return &res
}

func PBAllStoriesFromAllStories(as types.AllStories) *PBAllStories {
	res := &PBAllStories{}
	res.Data = map[int64]*PBStory{}

	for k, v := range as {
		res.Data[int64(k)] = PBStoryFromStory(v)
	}

	return res
}

func AllStoriesFromPBAllStories(ps *PBAllStories) (res types.AllStories) {
	res = types.AllStories{}

	for k, v := range ps.Data {
		res[int(k)] = StoryFromPBStory(v)
	}

	return res
}
