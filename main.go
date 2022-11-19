package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

type Brick struct {
	renderer *sdl.Renderer
	color *sdl.Color
	rect *sdl.Rect
	set *Bricks //the set of bricks the brick is part of
	hits int //how many hits it has left before it disappears
	//texture *some texture file*
	//index int //the brick's index in the array
}

//push draw functions to the queue
func (b *Brick) Draw() {
	b.renderer.SetDrawColor(b.color.R, b.color.G, b.color.G, b.color.A)
	b.renderer.FillRect(b.rect)
}

//pop draw functions from the queue
func (b *Brick) NoDraw(rq *RenderQueue) {
	//rq.Pop(b.index)
}

//remove brick from set
func (b *Brick) Break() {
	b.set.Pop(b.GetIndex())
}

//return the index of the brick
func (b Brick) GetIndex() int {
	for i := range b.set.bricks {
		if &b == b.set.bricks[i] {
			return i
		}
	}

	//hopefully this will never return 
	return len(b.set.bricks) - 1
}

type Bricks struct {
	bricks []*Brick
}

func (b *Bricks) Generate() {
	//var bArr []Brick
	for x := 0; x < 10; x++ {
		for y := 0; y < 3; y++ {
			m := &Brick{
				rect: &sdl.Rect{X: int32(x*40) + 120, Y: int32(y*20), W:40, H:20},
				color: &sdl.Color{R: CG()},
				set: b,
				hits: 1,
			}
			b.bricks = append(b.bricks, m)
		}
	}
}

func (b *Bricks) Pop(i int) {
	rm := func(i int) []*Brick {
		return append(b.bricks[:i], b.bricks[i+1:]...)
	}

	b.bricks = rm(i)
}

func (b *Bricks) Render() {
	for i := range b.bricks {
		b.bricks[i].Draw()
	}
}

func (bricks *Bricks) SetRenderer(ren *sdl.Renderer) {
	for i := range bricks.bricks {
		bricks.bricks[i].renderer = ren
	}
}

//male reproductive organ
type Balls struct {
	balls []*Ball
}

func (b *Balls) New() {
	b.balls = append(b.balls, &Ball{
		rect: &sdl.Rect{
			X: 0,
			Y: 440,
			W: 10,
			H: 10,
		},

		set: b,
		MaxVel: 10,
		MinVel: -10,
		isStuck: true,
	})
}

func (b *Balls) SetRenderer(ren *sdl.Renderer) {
	for i := range b.balls {
		b.balls[i].renderer = ren
	}
}

type Ball struct {
	renderer *sdl.Renderer
	rect *sdl.Rect
	set *Balls
	VelX float32
	VelY float32
	MaxVel float32
	MinVel float32
	isSticky bool
	isStuck bool
	fq []func()
}

func (ball *Ball) Draw() {
	ball.renderer.SetDrawColor(0,255,0,255)
	ball.renderer.FillRect(ball.rect)
}

//what the ball does when it collides with a brick
func (b *Ball) SingleCollide(brick *Brick) (int, bool) {
	if b.rect.HasIntersection(brick.rect) {
		//save the balls velocity
		tx, ty := b.VelX, b.VelY
			
		//stop the ball
		b.VelX = 0
		b.VelY = 0
		
		intersect, _ := b.rect.Intersect(brick.rect)

		//change direction on collision
		if intersect.H < intersect.W {
			b.VelX = tx
			b.VelY = -ty
		} else {
			b.VelX = -tx
			b.VelY = ty
		}

		//functions to call on collision
		for i := range b.fq {
			b.fq[i]()
		}

		//b.Split()
		
		return brick.GetIndex(), true
	}

	return 0, false
}

func (ball *Ball) CheckBoundaryCollide() {
		
	//side walls
	if ball.rect.X + ball.rect.W > 640 || ball.rect.X < 0 {
		ball.VelX = -ball.VelX
	}

	//ceiling
	if ball.rect.Y <= 0 {
		ball.rect.Y = 0
		ball.VelY = -ball.VelY
	}
}

func (ball *Ball) CapVelocity() {
	//cap velocity
	if ball.VelX > ball.MaxVel || ball.MinVel > ball.VelX {
		if ball.VelX > 0 {
			ball.VelX = ball.MaxVel
		} else {
			ball.VelX = ball.MinVel
		}
	}
}

func (ball *Ball) Split() {
	rect := &sdl.Rect{
		X: ball.rect.X,
		Y: ball.rect.Y,
		W: ball.rect.W,
		H: ball.rect.H,
	}
	VelX := ball.VelX
	VelY := ball.VelY
	set := ball.set
	MaxVel := ball.MaxVel
	MinVel := ball.MinVel
	fq := ball.fq
	
	ball.set.balls = append(ball.set.balls, &Ball{
		rect: rect,
		VelX: VelY,
		VelY: VelX,
		set: set,
		//isSticky: isSticky,
		MaxVel: MaxVel,
		MinVel: MinVel,
		fq: fq,
	})
}

//add a function to the array, return the index where it's stored
func (b *Ball) AddFunc(f func()) int {
	b.fq = append(b.fq, f)
	return len(b.fq) - 1
}

//remove a function from the array
func (b *Ball) PopFunc(i int) int {
	remove := func(s []func(), i int) []func() {
		return append(s[:i], s[i+1:]...)
	}

	b.fq = remove(b.fq, i)
	return i - 1
}

//what is a paddle but a rectangle?
type Paddle struct {
	rect *sdl.Rect
}

func (paddle *Paddle) UpdatePosition() {
	x, _, _ := sdl.GetMouseState()
	paddle.rect.X = x
}

//what the paddle does when it collides with the walls
func (paddle *Paddle) BoundaryCollide() {
	//add walls to the edges of the window
	if paddle.rect.X + paddle.rect.W > 640 {
		paddle.rect.X = 640 - paddle.rect.W
	}
}

//what the paddle does when it collides with the balls (or what the balls do)
func (paddle *Paddle) BallCollide(ball *Ball) {
	//maybe should make this its own function
	if ball.rect.HasIntersection(paddle.rect) {
		if ball.isSticky {
			//stop ball from moving on Y axis
			//ball should stick where it lands on paddle
			ball.rect.Y = paddle.rect.Y - ball.rect.H
			ball.isStuck = true
		}

		ball.VelY = -ball.VelY
		ball.VelX = -ball.VelX
	}

	//if the ball gets re-stuck, it should stick where it landed on the paddle
	if ball.isStuck {
		ball.rect.X = (paddle.rect.X + paddle.rect.W / 2) - (ball.rect.W / 2)
	} else {
		ball.rect.Y = ball.rect.Y - int32(ball.VelY)
		ball.rect.X = ball.rect.X - int32(ball.VelX)
	}
}

type RenderQueue struct {
	Arr []func()
	Ren *sdl.Renderer
}

//push a function to the end of the queue
func (r *RenderQueue) Push(f func()) {
	r.Arr = append(r.Arr, f)
}

//pop a function by index
func (r *RenderQueue) Pop(i int) {
	remove := func(s []func(), i int) []func() {
		return append(s[:i], s[i+1:]...)
	}

	remove(r.Arr, i)
}

func (r *RenderQueue) Flush() {
	r.Arr = []func(){}
}

//Color Gen
func CG() uint8 {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return uint8(r1.Intn(255))
}

func main() {
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		log.Fatal(err)
	}

	window, err := sdl.CreateWindow("Bruh", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 640, 480, sdl.WINDOW_SHOWN)
	if err != nil {
		log.Fatal(err)
	}

	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		log.Fatal(err)
	}

	defer renderer.Destroy()

	//create renderer function queue
	rq := &RenderQueue{
		Ren: renderer,
	}

	bricks := &Bricks{}
	bricks.Generate()
	bricks.SetRenderer(renderer)

	paddle := &Paddle{
		rect: &sdl.Rect{
			X: 0,
			Y: 450,
			W: 50,
			H: 10,
		},
	}

	balls := &Balls{}
	balls.New()
	balls.SetRenderer(renderer)
	balls.balls[0].Draw()

	sdl.ShowCursor(0)
	running := true
	
	for running {
		//initialize variable ; condition to check for ; value of variable after loop
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent:
				switch e.Keysym.Sym {
				case sdl.K_SPACE:
					balls.balls[0].isStuck = false
					balls.balls[0].VelY = 4
				case sdl.K_p:
					
				}
			}
		}

		//draw a black background
		renderer.SetDrawColor(0,0,0,0)
		renderer.Clear()

		//zip through the queue and call every func
		for i := range rq.Arr {
			rq.Arr[i]()
		}

		//for every ball, do
		for i := range balls.balls {
			balls.balls[i].Draw()
			balls.balls[i].CheckBoundaryCollide()
			balls.balls[i].CapVelocity()
			paddle.BallCollide(balls.balls[i])
		} //i should turn this into an array i can push and pop functions from

		for i := range bricks.bricks {
			var yes bool
			bricks.bricks[i].Draw()
			
			for j := range balls.balls {
				_, yes = balls.balls[j].SingleCollide(bricks.bricks[i])
			}

			if yes {
				bricks.bricks[i].Break()
				fmt.Println(bricks.bricks[i].GetIndex())
				break
			}
		}

		paddle.UpdatePosition()
		paddle.BoundaryCollide()
	
		//if you shorten the array while its currently looping
		//you will have an index out of range error


		//gonna need a object oriented way to do this

		renderer.SetDrawColor(255,255,255,255)
		renderer.FillRect(paddle.rect)
		
		renderer.Present()
		sdl.Delay(16)
	}
}
