package gsmutils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceRunesToGSM7(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{
			"＠￡＄￥ÇǾǿǺǻδ¯＿ɸφϕᵩƔɣγᴦᵧƛλᴧɷΏωώὠὡὢὣὤὥὦὧὨὩὪὫὬὭὮὯὼώᾠᾡᾢᾣᾤᾥᾦᾧᾨᾩᾪᾫᾬᾭᾮᾯῲῳῴῶῷῺΏῼπϖᴨψᴪςσθϑϴξǢǼᴁǣǽẞ！«»＂＃％＆＇`´｀（｟）｠＊＋，¸､－．·｡／０１¹２²３³４５６７８９：；＜＝＞？＾｛｝＼［～］｜¦￤￨",
			"@£$¥çØøÅåΔ__ΦΦΦΦΓΓΓΓΓΛΛΛΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΩΠΠΠΨΨΣΣΘΘΘΞÆÆÆææß!\"\"\"#%&''''(())*+,,,-.../0112233456789:;<=>?^{}\\[~]||||",
		},
		{
			"\"Kevin’s Massive Insanely Cool Birthday\": Hey friends! Please choose going if possible, they’re a little annoying about the group coming on time and I wanna get a good idea. Also, just wanna get a headcount for gift expectations!",
			"\"Kevin's Massive Insanely Cool Birthday\": Hey friends! Please choose going if possible, they're a little annoying about the group coming on time and I wanna get a good idea. Also, just wanna get a headcount for gift expectations!",
		},
	}
	for _, c := range cases {
		result := ReplaceRunesToGSM7(c.in)

		assert.Equal(t, c.out, result, fmt.Sprintf("testing => %+v", c.in))
	}
}
