package core_test

import (
	"context"
	"testing"
	"time"

	sloggorm "github.com/orandin/slog-gorm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	_ "github.com/BeamFoundry/rays-orm/internal/autoconfig"
	. "github.com/BeamFoundry/rays-orm/pkg/core"
	. "github.com/BeamFoundry/rays/pkg/core"
	. "github.com/BeamFoundry/rays/pkg/lang"
	. "github.com/BeamFoundry/rays/pkg/testing"
)

func TestOrm(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pkg/ORM Suite")
}

var _ = BeforeEach(func() {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{
		Logger: sloggorm.New(
			sloggorm.WithHandler(GetLogger().Handler()),
		),
	})
	Expect(err).NotTo(HaveOccurred())
	Expect(db.AutoMigrate(&User{})).NotTo(HaveOccurred())
	Expect(db.AutoMigrate(&Animal{})).NotTo(HaveOccurred())
	AddBeamFor(db)
	GetContainer().EnableScanning()
})

//nolint:govet // Entity
type Animal struct {
	ID     uint
	Name   string
	UserID uint
}

type AnimalRepository struct {
	OrmRepository[Animal, uint]
}

//nolint:govet // Entity
type User struct {
	ID        uint
	Name      string
	UserName  string
	Age       uint
	Animals   []Animal
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Foo struct{}

type UserRepositoryInterface interface {
	OrmRepositoryInterface[User, uint]
	FindByName(string) *User
	FindByUserName(string) *User
	FindWithErrorByUserName(string) (*User, Error)
	// FindByUserNameWithAnimals(string) *User
	// FindByUserNameWithAnimalsOrderByName(string) *User
	FindAllByName(string) []User
	FindOneByName(string) Optional[*User]
	FindAllByAgeIn([]uint) []User
	FindAllByUserNameAndAge(string, uint) []User
	FindAllByAgeBetweenOrderByAge(uint, uint) []User
	FindAllByAgeBetweenOrderByAgeAsc(uint, uint) []User
	FindAllByAgeBetweenOrderByAgeDesc(uint, uint) []User
}

var _ UserRepositoryInterface = &UserRepository{}

//nolint:unused // Testing
type UserRepository struct {
	OrmRepository[User, uint]
	findByName              func(string) *User
	findByUserName          func(string) *User
	findWithErrorByUserName func(string) (*User, Error)
	// findByUserNameWithAnimals               func(string) *User
	// findByUserNameWithAnimalsOrderByName    func(string) *User
	findAllByName                           func(string) []User
	findOneByName                           func(string) Optional[*User]
	findAllByAgeIn                          func([]uint) []User
	findAllByUserNameAndAge                 func(string, uint) []User
	findAllByAgeBetweenOrderByAge           func(uint, uint) []User
	findAllByAgeBetweenOrderByAgeAsc        func(uint, uint) []User
	findAllByAgeBetweenOrderByAgeDesc       func(uint, uint) []User
	findAllByAgeBetweenOrderByNameAscAgeAsc func(uint, uint) []User
	//
	findByNameNot                  func(string) []User
	findByNameIsNull               func() []User
	findByNameIsNotNull            func() []User
	findByNameOrName               func(string, string) []User
	findByNameLike                 func(string) []User
	findByNameNotLike              func(string) []User
	findAllByAgeBetween            func(uint, uint, ...OrmQueryFunc) []User
	queryAllByAgeBetween           func(uint, uint) *gorm.DB
	findAllByAgeNotIn              func([]uint) []User
	findAllByAgeGreaterThan        func(uint) []User
	findAllByAgeGreaterThanOrEqual func(uint) []User
	findAllByAgeLessThan           func(uint) []User
	findAllByAgeLessThanOrEqual    func(uint) []User
	//
	detectAllByName             func(string) []User
	findBadName                 func(string) []User
	findBadByNameIsHave         func(string) []User
	findBadByNameAndAge         func(string) []User
	findBadByNameOrderByAgeDown func(string) []User
	//
	findBadByFunction    func(string) *User
	findBadPointerByName func(string) *Foo
	findBadStructByName  func(string) Foo
	findBadSliceByName   func(string) []Foo
	findBadElemByName    func(string) string
	// findDistinctByName    func(string) []User
	findByNameOrderByFoo  func(string) []User
	findBadByNameWithFoos func(string) *User
	findByFoo             func(string) *User
	findBadByAge          func([]uint) []User
	findBadByAgeIn        func([]string) []User
	findBadTypeByName     func(uint) *User
}

func (this *UserRepository) FindByName(name string) *User {
	return this.findByName(name)
}

func (this *UserRepository) FindByUserName(name string) *User {
	return this.findByUserName(name)
}

func (this *UserRepository) FindWithErrorByUserName(name string) (*User, Error) {
	return this.findWithErrorByUserName(name)
}

/*
func (this *UserRepository) FindByUserNameWithAnimals(name string) *User {
	return this.findByUserNameWithAnimals(name)
}

func (this *UserRepository) FindByUserNameWithAnimalsOrderByName(name string) *User {
	return this.findByUserNameWithAnimalsOrderByName(name)
}
*/

func (this *UserRepository) FindAllByName(name string) []User {
	return this.findAllByName(name)
}

func (this *UserRepository) FindOneByName(name string) Optional[*User] {
	return this.findOneByName(name)
}

func (this *UserRepository) FindAllByAgeIn(age []uint) []User {
	return this.findAllByAgeIn(age)
}

func (this *UserRepository) FindAllByUserNameAndAge(name string, ages uint) []User {
	return this.findAllByUserNameAndAge(name, ages)
}

func (this *UserRepository) FindAllByAgeBetweenOrderByAge(start, end uint) []User {
	return this.findAllByAgeBetweenOrderByAge(start, end)
}

func (this *UserRepository) FindAllByAgeBetweenOrderByAgeAsc(start, end uint) []User {
	return this.findAllByAgeBetweenOrderByAgeAsc(start, end)
}

func (this *UserRepository) FindAllByAgeBetweenOrderByAgeDesc(start, end uint) []User {
	return this.findAllByAgeBetweenOrderByAgeDesc(start, end)
}

var _ = Describe("UserRepository", func() {
	var userRepository *UserRepository
	var animalRepository *AnimalRepository
	BeforeEach(func() {
		userRepository = &UserRepository{}
		animalRepository = &AnimalRepository{}

		AddBeamFor(userRepository)
		AddBeamFor(animalRepository)
		err := GetContainer().Initialize()
		Expect(err).NotTo(HaveOccurred())

		user, err := userRepository.Save(&User{Name: "Dorothy", UserName: "dorothy", Age: 25})
		Expect(err).NotTo(HaveOccurred())
		Expect(user).NotTo(BeNil())
		Expect(user.ID).NotTo(BeNumerically("==", 0))

		animal, err := animalRepository.Save(&Animal{Name: "Toto", UserID: user.ID})
		Expect(err).NotTo(HaveOccurred())
		Expect(animal).NotTo(BeNil())
		Expect(animal.ID).NotTo(BeNumerically("==", 0))
	})

	It("can find a record by its ID", func() {
		optional := userRepository.FindById(1)
		u, ok := optional.Get()
		Expect(ok).To(BeTrue())
		Expect(u.ID).To(BeNumerically("==", 1))
	})

	It("can find a record", func() {
		u := userRepository.FindByUserName("dorothy")
		Expect(u.ID).To(BeNumerically("==", 1))
		Expect(u.UserName).To(Equal("dorothy"))
	})

	It("can find a record with Error", func() {
		u, err := userRepository.FindWithErrorByUserName("dorothy")
		Expect(err).NotTo(HaveOccurred())
		Expect(u.ID).To(BeNumerically("==", 1))
		Expect(u.UserName).To(Equal("dorothy"))
		_, err = userRepository.FindWithErrorByUserName("toto")
		Expect(err).To(HaveOccurred())
	})

	/*
		It("can find a record with preload", func() {
			u := userRepository.FindByUserNameWithAnimals("dorothy")
			Expect(u.ID).To(BeNumerically("==", 1))
			Expect(u.UserName).To(Equal("dorothy"))
			Expect(u.Animals).NotTo(BeEmpty())
		})

		It("can find a record with preload and order", func() {
			u := userRepository.FindByUserNameWithAnimalsOrderByName("dorothy")
			Expect(u.ID).To(BeNumerically("==", 1))
			Expect(u.UserName).To(Equal("dorothy"))
			Expect(u.Animals).NotTo(BeEmpty())
		})
	*/

	It("can find many records", func() {
		users := userRepository.FindAllByName("Dorothy")
		Expect(users).NotTo(BeEmpty())
		u := users[0]
		Expect(u.ID).To(BeNumerically("==", 1))
		Expect(u.Name).To(Equal("Dorothy"))
	})

	It("can find optionally a record", func() {
		optional := userRepository.FindOneByName("Dorothy")
		u, ok := optional.Get()
		Expect(ok).To(BeTrue())
		Expect(u.ID).To(BeNumerically("==", 1))

		_, ok = userRepository.FindOneByName("Alice").Get()
		Expect(ok).To(BeFalse())
	})

	It("can find many records with In", func() {
		u := []*User{
			{Name: "Wendy", UserName: "wendy", Age: 30},
			{Name: "Alice", UserName: "alice", Age: 35},
		}
		_, err := userRepository.SaveAll(u)
		Expect(err).NotTo(HaveOccurred())
		users := userRepository.FindAllByAgeIn([]uint{25, 30})
		Expect(users).To(HaveLen(2))
	})

	It("can find many records with Between", func() {
		u := []*User{
			{Name: "Wendy", UserName: "wendy", Age: 30},
			{Name: "Alice", UserName: "alice", Age: 35},
		}
		_, err := userRepository.SaveAll(u)
		Expect(err).NotTo(HaveOccurred())
		users := userRepository.FindAllByAgeBetweenOrderByAge(20, 30)
		Expect(users).To(HaveLen(2))
	})

	It("can find many records and order them Desc", func() {
		u := []*User{
			{Name: "Wendy", UserName: "wendy", Age: 30},
			{Name: "Alice", UserName: "alice", Age: 35},
		}
		_, err := userRepository.SaveAll(u)
		Expect(err).NotTo(HaveOccurred())
		users := userRepository.FindAllByAgeBetweenOrderByAgeAsc(20, 30)

		Expect(users).To(HaveLen(2))
		Expect(users[0].Age).To(BeNumerically("==", 25))
	})

	It("can find many records and order them Desc", func() {
		u := []*User{
			{Name: "Wendy", UserName: "wendy", Age: 30},
			{Name: "Alice", UserName: "alice", Age: 35},
		}
		_, err := userRepository.SaveAll(u)
		Expect(err).NotTo(HaveOccurred())
		users := userRepository.FindAllByAgeBetweenOrderByAgeDesc(20, 30)

		Expect(users).To(HaveLen(2))
		Expect(users[0].Age).To(BeNumerically("==", 30))
	})

	It("can find many records and order by multiple fields", func() {
		u := []*User{
			{Name: "Wendy", UserName: "wendy", Age: 30},
			{Name: "Alice", UserName: "alice", Age: 35},
		}
		_, err := userRepository.SaveAll(u)
		Expect(err).NotTo(HaveOccurred())
		users := userRepository.findAllByAgeBetweenOrderByNameAscAgeAsc(20, 40)
		Expect(users).To(HaveLen(3))
		Expect(users[0].Name).To(Equal("Alice"))
		Expect(users[1].Name).To(Equal("Dorothy"))
		Expect(users[2].Name).To(Equal("Wendy"))
	})

	It("can find using is (not) null", func() {
		u := []*User{
			{Name: "Wendy", UserName: "wendy", Age: 30},
			{Name: "Alice", UserName: "alice", Age: 35},
		}
		_, err := userRepository.SaveAll(u)
		Expect(err).NotTo(HaveOccurred())
		users := userRepository.findByNameIsNotNull()
		Expect(users).To(HaveLen(3))
	})

	It("can find using or", func() {
		u := []*User{
			{Name: "Wendy", UserName: "wendy", Age: 30},
			{Name: "Alice", UserName: "alice", Age: 35},
		}
		_, err := userRepository.SaveAll(u)
		Expect(err).NotTo(HaveOccurred())
		users := userRepository.findByNameOrName("Dorothy", "Wendy")
		Expect(users).To(HaveLen(2))
	})

	It("can use variadic queries", func() {
		u := []*User{
			{Name: "Wendy", UserName: "wendy", Age: 30},
			{Name: "Alice", UserName: "alice", Age: 35},
		}
		_, err := userRepository.SaveAll(u)
		Expect(err).NotTo(HaveOccurred())
		order := func(db *gorm.DB) *gorm.DB {
			return db.Order("name DESC")
		}
		users := userRepository.findAllByAgeBetween(20, 30, order)
		Expect(users).To(HaveLen(2))
		Expect(users[0].Name).To(Equal("Wendy"))
		Expect(users[1].Name).To(Equal("Dorothy"))
	})

	It("can use return DB", func() {
		u := []*User{
			{Name: "Wendy", UserName: "wendy", Age: 30},
			{Name: "Alice", UserName: "alice", Age: 35},
		}
		_, err := userRepository.SaveAll(u)
		Expect(err).NotTo(HaveOccurred())
		var users []User
		userRepository.queryAllByAgeBetween(20, 30).Find(&users)
		Expect(users).To(HaveLen(2))
		var ages []uint
		userRepository.queryAllByAgeBetween(20, 30).Distinct("age").Order("age").Pluck("age", &ages)
		Expect(ages).To(HaveLen(2))
		Expect(ages[0]).To(BeNumerically("==", 25))
		Expect(ages[1]).To(BeNumerically("==", 30))
	})

	It("can query", func() {
		u := []*User{
			{Name: "Wendy", UserName: "wendy", Age: 30},
			{Name: "Alice", UserName: "alice", Age: 35},
		}
		_, err := userRepository.SaveAll(u)
		Expect(err).NotTo(HaveOccurred())

		var users []User
		ageBetween := func(from, to uint) OrmQueryFunc {
			return func(db *gorm.DB) *gorm.DB {
				return db.Where("age BETWEEN ? AND ?", from, to)
			}
		}
		userRepository.Query(ageBetween(20, 30)).Find(&users)
		Expect(users).To(HaveLen(2))
	})

	It("can update a record", func() {
		user, _ := userRepository.FindById(1).Get()
		user.Name = "foo"
		_, err := userRepository.Save(user)
		Expect(err).NotTo(HaveOccurred())

		opt := userRepository.FindById(1)
		u, ok := opt.Get()
		Expect(ok).To(BeTrue())
		Expect(u.Name).To(Equal("foo"))
	})

	It("can delete a record", func() {
		user, _ := userRepository.FindById(1).Get()
		Expect(userRepository.Delete(user)).NotTo(HaveOccurred())
		opt := userRepository.FindById(1)
		Expect(opt.IsPresent()).To(BeFalse())
	})

	It("can delete a record by its ID", func() {
		Expect(userRepository.DeleteById(1)).NotTo(HaveOccurred())
		opt := userRepository.FindById(1)
		Expect(opt.IsPresent()).To(BeFalse())
	})
})

type UserService struct {
	Service
	transactionManager GormTransactional       `@:"Inject"`
	userRepository     UserRepositoryInterface `@:"Inject"`
}

func (this *UserService) TransactionalChangeRollback() {
	this.transactionManager.Transactional(context.Background(), func(_ *gorm.DB) Error {
		user, _ := this.userRepository.FindById(1).Get()
		user.Name = "bar"
		_, err := this.userRepository.Save(user)
		Expect(err).NotTo(HaveOccurred())
		return NewError("TransactionalError", "Rolling back")
	})

	u := this.userRepository.FindByName("bar")
	Expect(u).To(BeNil())
}

func (this *UserService) TransactionalChange() {
	this.transactionManager.Transactional(context.Background(), func(_ *gorm.DB) Error {
		user, _ := this.userRepository.FindById(1).Get()
		user.Name = "bar"
		_, err := this.userRepository.Save(user)
		Expect(err).NotTo(HaveOccurred())
		return nil
	})

	u := this.userRepository.FindByName("bar")
	Expect(u).NotTo(BeNil())
	Expect(u.Name).To(Equal("bar"))
}

func (this *UserService) ErrorAfterRollback() {
	t := this.transactionManager.Start(context.Background())
	user, _ := this.userRepository.FindById(1).Get()
	t.Rollback()
	Expect(this.userRepository.Delete(user)).To(HaveOccurred())
	Expect(this.userRepository.DeleteById(1)).To(HaveOccurred())
	_, err := this.userRepository.Save(user)
	Expect(err).To(HaveOccurred())
	_, err = this.userRepository.SaveAll([]*User{user})
	Expect(err).To(HaveOccurred())
	this.transactionManager.End()
}

var _ = Describe("UserService", func() {
	var userRepository *UserRepository
	var userService *UserService
	BeforeEach(func() {
		userRepository = &UserRepository{}
		userService = &UserService{}

		AddBeamFor(userRepository)
		AddBeamFor(userService)
		err := GetContainer().Initialize()
		Expect(err).NotTo(HaveOccurred())

		user, err := userRepository.Save(&User{Name: "Dorothy", UserName: "dorothy", Age: 25})
		Expect(err).NotTo(HaveOccurred())
		Expect(user).NotTo(BeNil())
		Expect(user.ID).NotTo(BeNumerically("==", 0))
	})

	It("can start a Transaction and rollback it", func() {
		userService.TransactionalChangeRollback()
	})

	It("can start a Transaction and commit it", func() {
		userService.TransactionalChange()
	})
	It("can return an error if an operation is done after Rollback", func() {
		userService.ErrorAfterRollback()
	})
})
