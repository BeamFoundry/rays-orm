package util_test

import (
	"errors"
	"math"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	. "github.com/BeamFoundry/rays-orm/pkg/util"
	. "github.com/BeamFoundry/rays/pkg/lang"
)

func TestUtil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pkg/Util Suite")
}

//nolint:govet // Entity for testing
type Animal struct {
	ID     uint
	Name   string
	UserID uint
}

//nolint:govet // Entity for testing
type User struct {
	ID        uint
	Name      string
	UserName  string
	Age       uint
	Animals   []Animal
	DeletedAt gorm.DeletedAt
	CreatedAt time.Time
	UpdatedAt time.Time
}

var _ = Describe("OrmQueryFunc Utilities", func() {
	var db *gorm.DB

	BeforeEach(func() {
		var err error
		db, err = gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		sqlDB, err := db.DB()
		Expect(err).NotTo(HaveOccurred())
		sqlDB.SetMaxOpenConns(1)
		Expect(db.AutoMigrate(&User{}, &Animal{})).NotTo(HaveOccurred())

		users := []User{
			{Name: "Alice", UserName: "alice", Age: 20},
			{Name: "Bob", UserName: "bob", Age: 30},
			{Name: "Charlie", UserName: "charlie", Age: 40},
			{Name: "Alice", UserName: "alice2", Age: 25},
		}
		Expect(db.Create(&users).Error).NotTo(HaveOccurred())

		animal := Animal{Name: "Fluffy", UserID: users[0].ID}
		Expect(db.Create(&animal).Error).NotTo(HaveOccurred())
	})

	It("paginates results correctly", func() {
		var page1 []User
		Expect(db.Scopes(Paginate(1, 2), OrderByAsc("id")).Find(&page1).Error).NotTo(HaveOccurred())
		Expect(page1).To(HaveLen(2))
		Expect(page1[0].UserName).To(Equal("alice"))
		Expect(page1[1].UserName).To(Equal("bob"))

		var page2 []User
		Expect(db.Scopes(Paginate(2, 2), OrderByAsc("id")).Find(&page2).Error).NotTo(HaveOccurred())
		Expect(page2).To(HaveLen(2))
		Expect(page2[0].UserName).To(Equal("charlie"))
		Expect(page2[1].UserName).To(Equal("alice2"))

		// Page <= 0 defaults to 1, pageSize <= 0 defaults to DefaultPageSize
		var pageDefault []User
		Expect(db.Scopes(Paginate(0, 0), OrderByAsc("id")).Find(&pageDefault).Error).NotTo(HaveOccurred())
		Expect(pageDefault).To(HaveLen(4))
	})

	It("supports concurrent Paginate usage without race conditions", func() {
		fn := Paginate(0, 0)
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer GinkgoRecover()
				defer wg.Done()
				var users []User
				Expect(db.Scopes(fn).Find(&users).Error).NotTo(HaveOccurred())
				Expect(users).To(HaveLen(4))
			}()
		}
		wg.Wait()
	})

	It("returns an error on pagination offset overflow", func() {
		var users []User
		err := db.Scopes(Paginate(math.MaxInt, 10)).Find(&users).Error
		Expect(err).To(HaveOccurred())

		var langErr Error
		Expect(errors.As(err, &langErr)).To(BeTrue())
		Expect(langErr.Name()).To(Equal("PaginationOverflowError"))
		Expect(langErr.Message()).To(Equal("pagination offset overflow"))

		err2 := db.Scopes(Paginate(math.MaxInt/2+2, 2)).Find(&users).Error
		Expect(err2).To(HaveOccurred())
		var langErr2 Error
		Expect(errors.As(err2, &langErr2)).To(BeTrue())
		Expect(langErr2.Name()).To(Equal("PaginationOverflowError"))
	})

	It("does not leave reusable db handle in error state on pagination offset overflow", func() {
		fn := Paginate(math.MaxInt, 10)
		tx := fn(db)
		Expect(tx.Error).To(HaveOccurred())
		Expect(db.Error).ToNot(HaveOccurred())

		// A subsequent query using db should still work
		var users []User
		Expect(db.Find(&users).Error).NotTo(HaveOccurred())
		Expect(users).To(HaveLen(4))
	})

	It("limits results with Limit and WithLimit", func() {
		var users []User
		Expect(db.Scopes(Limit(2)).Find(&users).Error).NotTo(HaveOccurred())
		Expect(users).To(HaveLen(2))

		var usersWithLimit []User
		Expect(db.Scopes(WithLimit(2)).Find(&usersWithLimit).Error).NotTo(HaveOccurred())
		Expect(usersWithLimit).To(HaveLen(2))
	})

	It("enforces zero rows when combining scopes with WithLimit(0)", func() {
		var users []User
		Expect(db.Scopes(WithLimit(10), WithLimit(0)).Find(&users).Error).NotTo(HaveOccurred())
		Expect(users).To(BeEmpty())

		var usersWithOffset []User
		Expect(db.Scopes(WithOffset(1), WithLimit(10), WithLimit(0)).Find(&usersWithOffset).Error).NotTo(HaveOccurred())
		Expect(usersWithOffset).To(BeEmpty())

		var usersOverrideZero []User
		Expect(db.Scopes(WithLimit(0), WithLimit(2)).Find(&usersOverrideZero).Error).NotTo(HaveOccurred())
		Expect(usersOverrideZero).To(HaveLen(2))
	})

	It("offsets results with WithOffset", func() {
		var users []User
		Expect(db.Scopes(WithOffset(2), OrderByAsc("id")).Find(&users).Error).NotTo(HaveOccurred())
		Expect(users).To(HaveLen(2))
		Expect(users[0].UserName).To(Equal("charlie"))
	})

	It("cancels existing offset when zero offset is requested in WithOffset and Paginate", func() {
		var users []User
		// WithOffset(0) cancels prior offset
		Expect(db.Offset(2).Scopes(WithOffset(0), OrderByAsc("id")).Find(&users).Error).NotTo(HaveOccurred())
		Expect(users).To(HaveLen(4))
		Expect(users[0].UserName).To(Equal("alice"))

		// Paginate(1, 2) cancels prior offset
		var page1 []User
		Expect(db.Offset(2).Scopes(Paginate(1, 2), OrderByAsc("id")).Find(&page1).Error).NotTo(HaveOccurred())
		Expect(page1).To(HaveLen(2))
		Expect(page1[0].UserName).To(Equal("alice"))
		Expect(page1[1].UserName).To(Equal("bob"))

		// Paginate(2, 2) sets new nonzero offset
		var page2 []User
		Expect(db.Offset(1).Scopes(Paginate(2, 2), OrderByAsc("id")).Find(&page2).Error).NotTo(HaveOccurred())
		Expect(page2).To(HaveLen(2))
		Expect(page2[0].UserName).To(Equal("charlie"))
		Expect(page2[1].UserName).To(Equal("alice2"))
	})

	It("preloads associations", func() {
		var user User
		Expect(db.Scopes(WithPreload("Animals")).First(&user, "user_name = ?", "alice").Error).NotTo(HaveOccurred())
		Expect(user.Animals).To(HaveLen(1))
		Expect(user.Animals[0].Name).To(Equal("Fluffy"))
	})

	It("joins associations with WithJoins", func() {
		var users []User
		Expect(db.Scopes(WithJoins("JOIN animals ON animals.user_id = users.id")).Find(&users).Error).NotTo(HaveOccurred())
		Expect(users).To(HaveLen(1))
		Expect(users[0].UserName).To(Equal("alice"))
	})

	It("orders results with WithOrder, OrderByAsc, and OrderByDesc", func() {
		var usersAsc []User
		Expect(db.Scopes(OrderByAsc("age")).Find(&usersAsc).Error).NotTo(HaveOccurred())
		Expect(usersAsc[0].Age).To(BeNumerically("==", 20))

		var usersDesc []User
		Expect(db.Scopes(OrderByDesc("age")).Find(&usersDesc).Error).NotTo(HaveOccurred())
		Expect(usersDesc[0].Age).To(BeNumerically("==", 40))

		var usersCustom []User
		Expect(db.Scopes(WithOrder("age DESC, user_name ASC")).Find(&usersCustom).Error).NotTo(HaveOccurred())
		Expect(usersCustom[0].Age).To(BeNumerically("==", 40))
	})

	It("selects specific fields with WithSelect", func() {
		var users []User
		Expect(db.Scopes(WithSelect("name")).Find(&users).Error).NotTo(HaveOccurred())
		Expect(users[0].Name).NotTo(BeEmpty())
		Expect(users[0].Age).To(BeNumerically("==", 0))
	})

	It("omits specific fields with WithOmit", func() {
		var users []User
		Expect(db.Scopes(WithOmit("age")).Find(&users).Error).NotTo(HaveOccurred())
		Expect(users[0].Name).NotTo(BeEmpty())
		Expect(users[0].Age).To(BeNumerically("==", 0))
	})

	It("retrieves distinct values with WithDistinct", func() {
		var names []string
		Expect(db.Model(&User{}).Scopes(WithDistinct("name")).Pluck("name", &names).Error).NotTo(HaveOccurred())
		Expect(names).To(HaveLen(3))
	})

	It("filters with WithWhere and WithOr", func() {
		var users []User
		Expect(db.Scopes(WithWhere("age = ?", 20), WithOr("age = ?", 30)).Find(&users).Error).NotTo(HaveOccurred())
		Expect(users).To(HaveLen(2))
	})

	It("groups and filters with WithGroup and WithHaving", func() {
		type Result struct {
			Name  string
			Count int
		}
		var results []Result
		Expect(db.Model(&User{}).
			Select("name, count(*) as count").
			Scopes(WithGroup("name"), WithHaving("count(*) > ?", 1)).
			Scan(&results).Error).NotTo(HaveOccurred())

		Expect(results).To(HaveLen(1))
		Expect(results[0].Name).To(Equal("Alice"))
		Expect(results[0].Count).To(Equal(2))
	})

	It("handles soft deletes with WithUnscoped", func() {
		var user User
		Expect(db.First(&user, "user_name = ?", "alice").Error).NotTo(HaveOccurred())
		Expect(db.Delete(&user).Error).NotTo(HaveOccurred())

		var normalUsers []User
		Expect(db.Find(&normalUsers).Error).NotTo(HaveOccurred())
		Expect(normalUsers).To(HaveLen(3))

		var allUsers []User
		Expect(db.Scopes(WithUnscoped()).Find(&allUsers).Error).NotTo(HaveOccurred())
		Expect(allUsers).To(HaveLen(4))
	})

	It("queries with WithTable", func() {
		var count int64
		Expect(db.Scopes(WithTable("users")).Count(&count).Error).NotTo(HaveOccurred())
		Expect(count).To(BeNumerically("==", 4))
	})

	It("combines multiple query functions with Combine", func() {
		combined := Combine(
			OrderByAsc("age"),
			Limit(2),
		)
		var users []User
		Expect(db.Scopes(combined).Find(&users).Error).NotTo(HaveOccurred())
		Expect(users).To(HaveLen(2))
		Expect(users[0].Age).To(BeNumerically("==", 20))
		Expect(users[1].Age).To(BeNumerically("==", 25))
	})
})
