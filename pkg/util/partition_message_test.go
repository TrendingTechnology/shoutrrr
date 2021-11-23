package util

import (
	"fmt"
	"strings"

	"github.com/containrrr/shoutrrr/pkg/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Partition Message", func() {
	Describe("creating a json payload", func() {
		limits := types.MessageLimit{
			ChunkSize:      2000,
			TotalChunkSize: 6000,
			ChunkCount:     10,
		}
		When("given a message that exceeds the max length", func() {
			When("not splitting by lines", func() {
				It("should return a payload with chunked messages", func() {

					items, _ := testPartitionMessage(42, limits, 100)

					Expect(len(items[0].Text)).To(Equal(1994))
					Expect(len(items[1].Text)).To(Equal(1999))
					Expect(len(items[2].Text)).To(Equal(205))
				})
				It("omit characters above total max", func() {
					items, _ := testPartitionMessage(62, limits, 100)

					Expect(len(items[0].Text)).To(Equal(1994))
					Expect(len(items[1].Text)).To(Equal(1999))
					Expect(len(items[2].Text)).To(Equal(1999))
					Expect(len(items[3].Text)).To(Equal(5))
				})
			})
			When("splitting by lines", func() {
				It("should return a payload with chunked messages", func() {
					batches := testMessageItemsFromLines(18, limits, 2)
					items := batches[0]

					Expect(len(items[0].Text)).To(Equal(200))
					Expect(len(items[8].Text)).To(Equal(200))
				})
				When("the message items exceed the limits", func() {
					It("should split items into multiple batches", func() {
						batches := testMessageItemsFromLines(21, limits, 2)

						for b, chunks := range batches {
							fmt.Fprintf(GinkgoWriter, "Batch #%v: (%v chunks)\n", b, len(chunks))
							for c, chunk := range chunks {
								fmt.Fprintf(GinkgoWriter, " - Chunk #%v: (%v runes)\n", c, len(chunk.Text))
							}
						}

						Expect(len(batches)).To(Equal(2))
					})
				})
				It("should trim characters above chunk size", func() {
					hundreds := 42
					repeat := 21
					batches := testMessageItemsFromLines(hundreds, limits, repeat)
					items := batches[0]

					Expect(len(items[0].Text)).To(Equal(limits.ChunkSize))
					Expect(len(items[1].Text)).To(Equal(limits.ChunkSize))
				})
			})
		})
	})
})

const hundredChars = "this string is exactly (to the letter) a hundred characters long which will make the send func error"

func testMessageItemsFromLines(hundreds int, limits types.MessageLimit, repeat int) (batches [][]types.MessageItem) {

	builder := strings.Builder{}

	ri := 0
	for i := 0; i < hundreds; i++ {
		builder.WriteString(hundredChars)
		ri++
		if ri == repeat {
			builder.WriteRune('\n')
			ri = 0
		}
	}

	batches = MessageItemsFromLines(builder.String(), limits)

	// maxChunkSize := Min(limits.ChunkSize, repeat*100)

	// totalChunks := 0
	// for _, batchChunks := range batches {
	// 	totalChunks += len(batchChunks)
	// }

	//expectedTotalChunkCount := hundreds / repeat
	//expectedChunkCountInFirstBatch := Min(limits.TotalChunkSize/maxChunkSize, Min(expectedTotalChunkCount, limits.ChunkCount))
	// Expect(len(batches[0])).To(Equal(expectedChunkCountInFirstBatch), "Chunk count in first batch")

	// Expect(totalChunks).To(Equal(expectedTotalChunkCount), "Total chunk count")

	return
}

func testPartitionMessage(hundreds int, limits types.MessageLimit, distance int) (items []types.MessageItem, omitted int) {
	builder := strings.Builder{}

	for i := 0; i < hundreds; i++ {
		builder.WriteString(hundredChars)
	}

	items, omitted = PartitionMessage(builder.String(), limits, distance)

	contentSize := Min(hundreds*100, limits.TotalChunkSize)
	expectedChunkCount := CeilDiv(contentSize, limits.ChunkSize-1)
	expectedOmitted := Max(0, (hundreds*100)-contentSize)

	Expect(omitted).To(Equal(expectedOmitted))
	Expect(len(items)).To(Equal(expectedChunkCount))

	return
}
