package resolver

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/lygo/graphql-go"
	"github.com/lygo/graphql-go/cmd/gql-gen-resolver/testdata"
	"strconv"
	"strings"
)

func New() testdata.SchemaResolver {
	initHuman()
	initDroids()
	initStarships()

	return &Resolver{}
}

type human struct {
	ID        graphql.ID
	Name      string
	Friends   []graphql.ID
	AppearsIn []string
	Height    float64
	Mass      int
	Starships []graphql.ID
}

var humans = []*human{
	{
		ID:        "1000",
		Name:      "Luke Skywalker",
		Friends:   []graphql.ID{"1002", "1003", "2000", "2001"},
		AppearsIn: []string{"NEWHOPE", "EMPIRE", "JEDI"},
		Height:    1.72,
		Mass:      77,
		Starships: []graphql.ID{"3001", "3003"},
	},
	{
		ID:        "1001",
		Name:      "Darth Vader",
		Friends:   []graphql.ID{"1004"},
		AppearsIn: []string{"NEWHOPE", "EMPIRE", "JEDI"},
		Height:    2.02,
		Mass:      136,
		Starships: []graphql.ID{"3002"},
	},
	{
		ID:        "1002",
		Name:      "Han Solo",
		Friends:   []graphql.ID{"1000", "1003", "2001"},
		AppearsIn: []string{"NEWHOPE", "EMPIRE", "JEDI"},
		Height:    1.8,
		Mass:      80,
		Starships: []graphql.ID{"3000", "3003"},
	},
	{
		ID:        "1003",
		Name:      "Leia Organa",
		Friends:   []graphql.ID{"1000", "1002", "2000", "2001"},
		AppearsIn: []string{"NEWHOPE", "EMPIRE", "JEDI"},
		Height:    1.5,
		Mass:      49,
	},
	{
		ID:        "1004",
		Name:      "Wilhuff Tarkin",
		Friends:   []graphql.ID{"1001"},
		AppearsIn: []string{"NEWHOPE"},
		Height:    1.8,
		Mass:      0,
	},
}

var humanData = make(map[graphql.ID]*human)

func initHuman() {
	for _, h := range humans {
		humanData[h.ID] = h
	}
}

type droid struct {
	ID              graphql.ID
	Name            string
	Friends         []graphql.ID
	AppearsIn       []string
	PrimaryFunction string
}

var droids = []*droid{
	{
		ID:              "2000",
		Name:            "C-3PO",
		Friends:         []graphql.ID{"1000", "1002", "1003", "2001"},
		AppearsIn:       []string{"NEWHOPE", "EMPIRE", "JEDI"},
		PrimaryFunction: "Protocol",
	},
	{
		ID:              "2001",
		Name:            "R2-D2",
		Friends:         []graphql.ID{"1000", "1002", "1003"},
		AppearsIn:       []string{"NEWHOPE", "EMPIRE", "JEDI"},
		PrimaryFunction: "Astromech",
	},
}

var droidData = make(map[graphql.ID]*droid)

func initDroids() {
	for _, d := range droids {
		droidData[d.ID] = d
	}
}

type starship struct {
	ID     graphql.ID
	Name   string
	Length float64
}

var starships = []*starship{
	{
		ID:     "3000",
		Name:   "Millennium Falcon",
		Length: 34.37,
	},
	{
		ID:     "3001",
		Name:   "X-Wing",
		Length: 12.5,
	},
	{
		ID:     "3002",
		Name:   "TIE Advanced x1",
		Length: 9.2,
	},
	{
		ID:     "3003",
		Name:   "Imperial shuttle",
		Length: 20,
	},
}

var starshipData = make(map[graphql.ID]*starship)

func initStarships() {
	for _, s := range starships {
		starshipData[s.ID] = s
	}
}

type review struct {
	stars      int32
	commentary *string
}

var reviews = make(map[string][]*review)

type Resolver struct{}

func (r *Resolver) Hero(ctx context.Context, in testdata.HeroArguments) (*testdata.CharacterResolver, error) {
	if in.Episode == testdata.Episode_EMPIRE {
		return &testdata.CharacterResolver{&humanResolver{humanData["1000"]}}, nil
	}
	return &testdata.CharacterResolver{&droidResolver{droidData["2001"]}}, nil
}

func (r *Resolver) Reviews(ctx context.Context, in testdata.ReviewsArguments) ([]testdata.ReviewResolver, error) {
	ep := in.Episode

	var l []testdata.ReviewResolver
	if ep == nil {
		for _, review := range reviews[testdata.Episode_NEWHOPE] {
			l = append(l, &reviewResolver{review})
		}
		for _, review := range reviews[testdata.Episode_EMPIRE] {
			l = append(l, &reviewResolver{review})
		}
		for _, review := range reviews[testdata.Episode_JEDI] {
			l = append(l, &reviewResolver{review})
		}
	} else {
		for _, review := range reviews[string(*ep)] {
			l = append(l, &reviewResolver{review})
		}
	}

	return l, nil
}

func (r *Resolver) Search(ctx context.Context, in testdata.SearchArguments) ([]*testdata.SearchResultResolver, error) {
	var l []*testdata.SearchResultResolver
	for _, h := range humans {
		if strings.Contains(h.Name, in.Text) {
			l = append(l, &testdata.SearchResultResolver{
				Result: &humanResolver{h},
			})
		}
	}
	for _, d := range droids {
		if strings.Contains(d.Name, in.Text) {
			l = append(l, &testdata.SearchResultResolver{
				Result: &droidResolver{d},
			})
		}
	}
	for _, s := range starships {
		if strings.Contains(s.Name, in.Text) {
			l = append(l, &testdata.SearchResultResolver{
				Result: &starshipResolver{s},
			})
		}
	}
	return l, nil
}

func (r *Resolver) Character(ctx context.Context, in testdata.CharacterArguments) (*testdata.CharacterResolver, error) {
	if h := humanData[in.ID]; h != nil {
		return &testdata.CharacterResolver{Character: &humanResolver{h: h}}, nil
	}
	if d := droidData[in.ID]; d != nil {
		return &testdata.CharacterResolver{Character: &droidResolver{d: d}}, nil
	}
	return nil, nil
}

func (r *Resolver) Droid(ctx context.Context, in testdata.DroidArguments) (testdata.DroidResolver, error) {
	if d := droidData[in.ID]; d != nil {
		return &droidResolver{d: d}, nil
	}
	return nil, nil
}

func (r *Resolver) Human(ctx context.Context, in testdata.HumanArguments) (testdata.HumanResolver, error) {
	if h := humanData[in.ID]; h != nil {
		return &humanResolver{h: h}, nil
	}
	return nil, nil
}

func (r *Resolver) Starship(ctx context.Context, in testdata.StarshipArguments) (testdata.StarshipResolver, error) {
	if s := starshipData[in.ID]; s != nil {
		return &starshipResolver{s: s}, nil
	}
	return nil, nil
}

func (r *Resolver) CreateReview(ctx context.Context, in testdata.CreateReviewArguments) (testdata.ReviewResolver, error) {
	review := &review{
		stars:      in.Review.Stars,
		commentary: in.Review.Commentary,
	}
	reviews[in.Episode] = append(reviews[in.Episode], review)
	return &reviewResolver{r: review}, nil
}

type humanResolver struct {
	h *human
}

var _ testdata.HumanResolver = &humanResolver{}

func (r *humanResolver) ID() graphql.ID {
	return r.h.ID
}

func (r *humanResolver) Name() string {
	return r.h.Name
}

func (r *humanResolver) Height(args testdata.HeightArguments) float64 {
	return convertLength(r.h.Height, args.Unit)
}

func (r *humanResolver) Mass() *float64 {
	if r.h.Mass == 0 {
		return nil
	}
	f := float64(r.h.Mass)
	return &f
}

func (r *humanResolver) Friends(ctx context.Context) (*[]*testdata.CharacterResolver, error) {
	return resolveCharacters(r.h.Friends), nil
}

func (r *humanResolver) FriendsConnection(ctx context.Context, in testdata.FriendsConnectionArguments) (testdata.FriendsConnectionResolver, error) {
	return newFriendsConnectionResolver(r.h.Friends, in)
}

func (r *humanResolver) AppearsIn(ctx context.Context) ([]testdata.Episode, error) {
	return r.h.AppearsIn, nil
}

func (r *humanResolver) Starships(ctx context.Context) (*[]testdata.StarshipResolver, error) {
	l := make([]testdata.StarshipResolver, len(r.h.Starships))
	for i, id := range r.h.Starships {
		l[i] = &starshipResolver{s: starshipData[id]}
	}
	return &l, nil
}

type droidResolver struct {
	d *droid
}

var _ testdata.DroidResolver = &droidResolver{}

func (r *droidResolver) ID() graphql.ID {
	return r.d.ID
}

func (r *droidResolver) Name() string {
	return r.d.Name
}

func (r *droidResolver) Friends(ctx context.Context) (*[]*testdata.CharacterResolver, error) {
	return resolveCharacters(r.d.Friends), nil
}

func (r *droidResolver) FriendsConnection(ctx context.Context, in testdata.FriendsConnectionArguments) (testdata.FriendsConnectionResolver, error) {
	return newFriendsConnectionResolver(r.d.Friends, in)
}

func (r *droidResolver) AppearsIn(ctx context.Context) ([]testdata.Episode, error) {
	return r.d.AppearsIn, nil
}

func (r *droidResolver) PrimaryFunction() *string {
	if r.d.PrimaryFunction == "" {
		return nil
	}
	return &r.d.PrimaryFunction
}

type starshipResolver struct {
	s *starship
}

var _ testdata.StarshipResolver = &starshipResolver{}

func (r *starshipResolver) ID() graphql.ID {
	return r.s.ID
}

func (r *starshipResolver) Name() string {
	return r.s.Name
}

func (r *starshipResolver) Length(args testdata.LengthArguments) float64 {
	return convertLength(r.s.Length, args.Unit)
}

type searchResultResolver struct {
	result interface{}
}

func (r *searchResultResolver) ToHuman() (*humanResolver, bool) {
	res, ok := r.result.(*humanResolver)
	return res, ok
}

func (r *searchResultResolver) ToDroid() (*droidResolver, bool) {
	res, ok := r.result.(*droidResolver)
	return res, ok
}

func (r *searchResultResolver) ToStarship() (*starshipResolver, bool) {
	res, ok := r.result.(*starshipResolver)
	return res, ok
}

func convertLength(meters float64, unit string) float64 {
	switch unit {
	case "METER":
		return meters
	case "FOOT":
		return meters * 3.28084
	default:
		panic("invalid unit")
	}
}

func resolveCharacters(ids []graphql.ID) *[]*testdata.CharacterResolver {
	var characters []*testdata.CharacterResolver
	for _, id := range ids {
		if c := resolveCharacter(id); c != nil {
			characters = append(characters, c)
		}
	}
	return &characters
}

func resolveCharacter(id graphql.ID) *testdata.CharacterResolver {
	if h, ok := humanData[id]; ok {
		return &testdata.CharacterResolver{Character: &humanResolver{h: h}}
	}
	if d, ok := droidData[id]; ok {
		return &testdata.CharacterResolver{Character: &droidResolver{d: d}}
	}
	return nil
}

type reviewResolver struct {
	r *review
}

var _ testdata.ReviewResolver = &reviewResolver{}

func (r *reviewResolver) Stars() int32 {
	return r.r.stars
}

func (r *reviewResolver) Commentary() *string {
	return r.r.commentary
}

type friendsConnectionResolver struct {
	ids  []graphql.ID
	from int
	to   int
}

var _ testdata.FriendsConnectionResolver = &friendsConnectionResolver{}

func newFriendsConnectionResolver(ids []graphql.ID, args testdata.FriendsConnectionArguments) (*friendsConnectionResolver, error) {
	from := 0
	if args.After != nil {
		b, err := base64.StdEncoding.DecodeString(string(*args.After))
		if err != nil {
			return nil, err
		}
		i, err := strconv.Atoi(strings.TrimPrefix(string(b), "cursor"))
		if err != nil {
			return nil, err
		}
		from = i
	}

	to := len(ids)
	if args.First != nil {
		to = from + int(*args.First)
		if to > len(ids) {
			to = len(ids)
		}
	}

	return &friendsConnectionResolver{
		ids:  ids,
		from: from,
		to:   to,
	}, nil
}

func (r *friendsConnectionResolver) TotalCount() int32 {
	return int32(len(r.ids))
}

func (r *friendsConnectionResolver) Edges(ctx context.Context) (*[]testdata.FriendsEdgeResolver, error) {
	l := make([]testdata.FriendsEdgeResolver, r.to-r.from)
	for i := range l {
		l[i] = &friendsEdgeResolver{
			cursor: encodeCursor(r.from + i),
			id:     r.ids[r.from+i],
		}
	}
	return &l, nil
}

func (r *friendsConnectionResolver) Friends(ctx context.Context) (*[]*testdata.CharacterResolver, error) {
	return resolveCharacters(r.ids[r.from:r.to]), nil
}

func (r *friendsConnectionResolver) PageInfo(ctx context.Context) (testdata.PageInfoResolver, error) {
	return &pageInfoResolver{
		startCursor: encodeCursor(r.from),
		endCursor:   encodeCursor(r.to - 1),
		hasNextPage: r.to < len(r.ids),
	}, nil
}

func encodeCursor(i int) graphql.ID {
	return graphql.ID(base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("cursor%d", i+1))))
}

type friendsEdgeResolver struct {
	cursor graphql.ID
	id     graphql.ID
}

var _ testdata.FriendsEdgeResolver = &friendsEdgeResolver{}

func (r *friendsEdgeResolver) Cursor() graphql.ID {
	return r.cursor
}

func (r *friendsEdgeResolver) Node(ctx context.Context) (*testdata.CharacterResolver, error) {
	return resolveCharacter(r.id), nil
}

type pageInfoResolver struct {
	startCursor graphql.ID
	endCursor   graphql.ID
	hasNextPage bool
}

var _ testdata.PageInfoResolver = &pageInfoResolver{}

func (r *pageInfoResolver) StartCursor() *graphql.ID {
	return &r.startCursor
}

func (r *pageInfoResolver) EndCursor() *graphql.ID {
	return &r.endCursor
}

func (r *pageInfoResolver) HasNextPage() bool {
	return r.hasNextPage
}
