package main

import (
	"fmt"
)

type storyPage struct {
	text string
	// nextPage storyPage // wouldn't work because of recursion
	nextPage *storyPage
}

func playStory(page *storyPage) {
	if page == nil {
		return
	}

	fmt.Println(page.text)
	playStory(page.nextPage)
}

// this function acts as a method for storyPage struct; the type declared is called receiver type
func (page *storyPage) playStoryMethod() {
	for page != nil {
		fmt.Println(page.text)
		page = page.nextPage
	}
}

func (page *storyPage) addToEnd(text string) {
	for page.nextPage != nil {
		page = page.nextPage
	}
	page.nextPage = &storyPage{text, nil}
}

func (page *storyPage) addAfter(text string) {
	newPage := &storyPage{text, page.nextPage}
	page.nextPage = newPage
}

func main() {
	// scanner := bufio.NewScanner(os.Stdin)

	page1 := storyPage{"It was a dark and stormy night.", nil}

	/*page2 := storyPage{"You are alone and you need to find a sacred helmet before the bad guys do.", nil}
	page3 := storyPage{"You see a troll ahead.", nil}

	page1.nextPage = &page2
	page2.nextPage = &page3 */

	page1.addToEnd("You are alone and you need to find a sacred helmet before the bad guys do.")
	page1.addAfter("...")
	page1.addToEnd("You see a troll ahead.")

	// playStory(&page1)
	page1.playStoryMethod()
}
