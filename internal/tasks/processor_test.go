package tasks_test

import (
	"context"
	"kosis/internal/config"
	"kosis/internal/db"
	"kosis/internal/models"
	"kosis/internal/tasks"
	"kosis/internal/testhelpers"
	"strings"

	"github.com/hibiken/asynq"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

var _ = Describe("HandleFetchReportsTask", func() {
	var dbConn *gorm.DB
	var p *tasks.TaskProcessor
	var listWithOneReport = `{
		"status": "000",
		"message": "정상",
		"page_no": 1,
		"page_count": 1,
		"total_count": 1,
		"total_page": 1,
		"list": [
			{
				"corp_code": "00356361",
				"corp_name": "LG화학",
				"stock_code": "051910",
				"corp_cls": "Y",
				"report_nm": "분기보고서 (2025.09)",
				"rcept_no": "20251114001374",
				"flr_nm": "LG화학",
				"rcept_dt": "20251114",
				"rm": ""
			}
		]
	}`

	var errRes1 = `{ "status": "014", "message": "문서가 존재하지 않습니다." }`
	var errRes2 = `{ "status": "010","message":"등록되지 않은 인증키입니다." }`
	var errRes3 = `{ "status": "013","message":"등록된 종목코드 또는 고유번호가 아닙니다." }`

	var testDocument = `<DOCUMENT>
				<CONTENT>
					<TEXT>
						<P>Hello, world!</P>
					</TEXT>
				</CONTENT>
			</DOCUMENT>`

	BeforeEach(func() {
		cfg, err := config.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		dbConn, err = db.InitDB(cfg.DatabaseURL)
		Expect(err).NotTo(HaveOccurred())

		testhelpers.CleanupDB(dbConn)

		p = tasks.NewTaskProcessor(dbConn, cfg)

		testhelpers.Activate()
		p.GetDartClient().UseDefaultClient()
	})

	AfterEach(func() {
		testhelpers.Deactivate()
	})

	It("stores raw reports", func() {
		testhelpers.New("https://opendart.fss.or.kr").
			Get("/api/list.json").Reply(200).
			BodyString(listWithOneReport).
			Header("Content-Type", "application/json")

		zipDocument, err := testhelpers.CreateMockZipArchive("document.xml", []byte(testDocument))
		Expect(err).NotTo(HaveOccurred())

		testhelpers.New("https://opendart.fss.or.kr").Get("/api/document.xml").Reply(200).Body(zipDocument).Header("Content-Type", "application/zip").Header("Content-Disposition", `attachment; filename="document.zip"`)

		ctx := context.Background()
		err = p.HandleFetchReportsTask(ctx, asynq.NewTask(tasks.TypeTaskFetchReports, []byte("{}")))
		Expect(err).NotTo(HaveOccurred())

		result, err := gorm.G[models.RawReport](dbConn).Where("corp_code = ?", "00356361").First(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.ReceiptNumber).To(Equal("20251114001374"))
		Expect(result.CorpCode).To(Equal("00356361"))
		Expect(strings.TrimSpace(string(result.BlobData))).To(Equal(strings.TrimSpace(testDocument)))
	})

	It("skips if raw report already exists", func() {
		testhelpers.New("https://opendart.fss.or.kr").
			Get("/api/list.json").Reply(200).BodyString(listWithOneReport).Header("Content-Type", "application/json")

		rawReport := models.RawReport{
			ReceiptNumber: "20251114001374",
			CorpCode:      "00356361",
			BlobData:      []byte("<DOCUMENT><CONTENT><TEXT><P>Hello, world!</P></TEXT></CONTENT></DOCUMENT>"),
			BlobSize:      41,
		}

		ctx := context.Background()
		result := gorm.WithResult()

		err := gorm.G[models.RawReport](dbConn, result).Create(ctx, &rawReport)
		Expect(err).NotTo(HaveOccurred())

		err = p.HandleFetchReportsTask(ctx, asynq.NewTask(tasks.TypeTaskFetchReports, []byte("{}")))
		Expect(err).NotTo(HaveOccurred())
	})

	DescribeTable("Handle errors from Dart API",
		func(bodyString string) {
			testhelpers.New("https://opendart.fss.or.kr").
				Get("/api/list.json").Reply(200).BodyString(bodyString)

			ctx := context.Background()
			err := p.HandleFetchReportsTask(ctx, asynq.NewTask(tasks.TypeTaskFetchReports, []byte("{}")))
			Expect(err).NotTo(HaveOccurred())

			count, err := gorm.G[models.RawReport](dbConn).Count(ctx, "id")
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(BeZero())
		},
		Entry("document not found", errRes1),
		Entry("invalid API key", errRes2),
		Entry("invalid corp code", errRes3),
	)
})
