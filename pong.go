package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

type color struct {
	r, g, b byte
}

var (
	bkg                                            sdl.Color
	title                                                = "SDL2"
	winWidth, winHeight                            int32 = 800, 800
	renderer                                       *sdl.Renderer
	window                                         *sdl.Window
	lastFrame, lastTime, frameCount, timerFPS, fps uint64
	setFPS                                         uint64 = 60
	mouse                                          sdl.Point
	mouseState                                     uint32
	keystate                                       = sdl.GetKeyboardState()
	event                                          sdl.Event
	running                                        bool
	padHeight                                      float32 = 100.0
	padWidth                                       float32 = 20.0
)

type position struct {
	x, y float32
}

type ball struct {
	position
	radius int32
	xv, yv float32
	color  sdl.Color
}

type paddles struct {
	position
	width  float32
	height float32
	color  sdl.Color
}

// we draw starting from top left to bottom right.
// paddle will be a rectangle where position is the center of it.
func (pad *paddles) draw() {
	// YAGNI - Ya Ain't Gonna Need It
	// going too deep into a thing might be a mistake, you might not need it in the future!

	// so we go "back" half of the width from the center
	// and "up" half of the height to reach top left.
	startX := pad.x - pad.width/2
	startY := pad.y - pad.height/2
	renderer.SetDrawColor(setRGBAColor(pad.color))

	// reason we start with y loop is cauze, imagine an array wich represents all the pixels on the screen.
	// The screen is a matrix a x b.
	// Since RAM is slower than CPU (around 100-200 cycles), and when you ask the ram for an array you get
	// a chunk of bytes for the next elements, it's much faster to iterate the array in order of asc index.
	//  _____
	// |     |
	// _______
	// |     |
	// ________   if you start with the x you will loop, like in a normal matrix, through the first elements of the rows.
	// |_____|    [0,0], [0,1], [0,2] ecc, which in the array would be distant a len(row) => hopping through the array
	//            requires asking ram for memory!!

	for y := 0; y < int(pad.height); y++ {
		for x := 0; x < int(pad.width); x++ {
			renderer.DrawPoint(int32(startX)+int32(x), int32(startY)+int32(y))
		}
	}
}

func (ball *ball) draw() {

	startX := -ball.radius
	startY := -ball.radius

	renderer.SetDrawColor(setRGBAColor(ball.color))

	for y := startX; y < ball.radius; y++ {
		for x := startY; x < ball.radius; x++ {

			// we draw pixels only if they are within the circle
			// avoid square root coz its expensive on the cpu!
			// so just do x^2+y^2 <= r^2
			if x*x+y*y <= ball.radius*ball.radius {
				renderer.DrawPoint(int32(ball.x)+x, int32(ball.y)+y)

			}
		}
	}

}

func (ball *ball) update(paddle1 paddles, paddle2 paddles) {
	ball.x += ball.xv
	ball.y += ball.yv

	// score a point
	if ball.x-float32(ball.radius) <= 0 || ball.x+float32(ball.radius) >= float32(winWidth) {
		ball.x = float32(winWidth) / 2
		return
	}

	// handle collisions
	// top of the screen ||bottom of screen
	if int32(ball.y)-ball.radius < 0 || int32(ball.y)+int32(ball.radius) > winHeight {
		ball.yv = -ball.yv
	}
	if int32(ball.x)-ball.radius < 0 || int32(ball.x)+int32(ball.radius) > winWidth {
		ball.xv = -ball.xv
	}

	// bounce off paddles
	if (ball.x-float32(ball.radius)) < paddle1.x+padWidth/2 && (ball.y+float32(ball.radius) > paddle1.y-padHeight/2 && ball.y+float32(ball.radius) < paddle1.y+padHeight/2) {
		ball.xv = -ball.xv
	}
	if (ball.x+float32(ball.radius)) > paddle2.x-padWidth/2 && (ball.y+float32(ball.radius) > paddle2.y-padHeight/2 && ball.y+float32(ball.radius) < paddle2.y+padHeight/2) {
		ball.xv = -ball.xv
	}

}

func (paddle *paddles) aiUpdate(ball *ball) {
	paddle.y = ball.y
}

func (paddle *paddles) update() {

	if keystate[sdl.SCANCODE_UP] != 0 {
		paddle.y -= 10
	}
	if keystate[sdl.SCANCODE_DOWN] != 0 {
		paddle.y += 10

	}

	if paddle.y-paddle.height/2 < 0 || paddle.y+paddle.height/2 > float32(winHeight) {
		paddle.y = float32(winHeight - 50)
	}

	paddle.draw()
}

func main() {
	start()

	pad1 := paddles{position{padWidth, float32(winHeight / 2)}, padWidth, padHeight, setSdlColor(255, 0, 255, 255)}
	pad2 := paddles{position{float32(winWidth) - padWidth, float32(winHeight / 2)}, padWidth, padHeight, setSdlColor(255, 0, 255, 255)}
	ball := ball{position{300, 300}, 20, 7, 10, setSdlColor(255, 0, 255, 255)}
	// here if you want to create new windows
	//startSet("hello", 300, 300)
	bkg = setSdlColor(0, 0, 0, 255)
	for running {
		daemon()
		//tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(winWidth), int32(winHeight))
		//if err != nil {
		//	panic(err)
		//}
		beginRender(nil)
		// Draw stuff here
		pad1.draw()
		pad2.draw()
		ball.draw()
		endRender()

		// escape key sets running state to false and quits window
		if keystate[sdl.SCANCODE_ESCAPE] != 0 {
			running = false
		}

		// Update screen here
		ball.update(pad1, pad2)
		pad1.update()
		pad2.aiUpdate(&ball)

	}
	quit()

}

func setSdlColor(r, g, b, a uint8) sdl.Color {
	var c sdl.Color
	c.R = r
	c.G = g
	c.B = b
	c.A = a
	return c
}

func setRGBAColor(color sdl.Color) (r, g, b, a uint8) {
	r = color.R
	g = color.G
	b = color.B
	a = color.A
	return
}

func start() {
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "0")
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	window, err := sdl.CreateWindow(title, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, winWidth, winHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		panic(err)
	}
	running = true
}

func startSet(t string, w, h int32) {
	title = t
	winWidth = w
	winHeight = h
	start()
}

// call this after main loop
func quit() {
	running = false
	window.Destroy()
	renderer.Destroy()
	sdl.Quit()
}

// daemon keeps track of frame stuff and calls input to check players input
func daemon() {
	lastFrame = sdl.GetTicks64()
	if lastFrame >= (lastTime + 1000) {
		lastTime = lastFrame
		fps = frameCount
		frameCount = 0
	}
	input()
}

func input() {

	// This gets keyboard state
	keystate = sdl.GetKeyboardState()

	// this for loops waits for quitevent
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch event.(type) {
		case *sdl.QuitEvent:
			running = false
			break

		}

		// to avoid CPU going crazy
	}
	mouse.X, mouse.Y, mouseState = sdl.GetMouseState()
}

// define textures in begin and endRender functions and set the render
// draw target to that
// Trying git commit
func beginRender(texture *sdl.Texture) {
	renderer.SetDrawColor(bkg.R, bkg.G, bkg.B, bkg.A)
	if texture != nil {
		renderer.SetRenderTarget(texture)
	}
	renderer.Clear()
	frameCount++
	timerFPS = sdl.GetTicks64() - lastFrame
	if timerFPS < (1000 / setFPS) {
		sdl.Delay(uint32((1000 / setFPS) - timerFPS))
	}
	//renderer.SetDrawColor(255, 0, 0, 255)
}

// if you want to swap a texture to the renderer itself, do it here
func endRender() {
	renderer.Present()
}
