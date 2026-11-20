package core_test

import (
	"context"
	"reflect"
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
	. "github.com/BeamFoundry/rays/pkg/util/stream"
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
	FindStreamByName(string) Stream[User]
	StreamAllByName(string) Stream[User]
	FindStreamWithErrorByName(string) (Stream[User], Error)
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
	findStreamByName                        func(string) Stream[User]
	streamAllByName                         func(string) Stream[User]
	findStreamWithErrorByName               func(string) (Stream[User], Error)
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
	findBadStreamByName  func(string) Stream[Foo]
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

func (this *UserRepository) FindStreamByName(name string) Stream[User] {
	return this.findStreamByName(name)
}

func (this *UserRepository) StreamAllByName(name string) Stream[User] {
	return this.streamAllByName(name)
}

func (this *UserRepository) FindStreamWithErrorByName(name string) (Stream[User], Error) {
	return this.findStreamWithErrorByName(name)
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
		Expect(err.Name()).To(Equal("GormError"))
		Expect(err).To(MatchError(gorm.ErrRecordNotFound))
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

	It("can find many records returning a Stream", func() {
		u := []*User{
			{Name: "Dorothy", UserName: "dorothy2", Age: 30},
			{Name: "Dorothy", UserName: "dorothy3", Age: 40},
		}
		_, err := userRepository.SaveAll(u)
		Expect(err).NotTo(HaveOccurred())

		st := userRepository.FindStreamByName("Dorothy")
		names := st.Filter(func(user User) bool { return user.Age >= 30 }).
			SortedBy(func(user User) int { return int(user.Age) }).
			Map(func(user User) string { return user.UserName }).
			ToSlice()
		Expect(names).To(Equal([]string{"dorothy2", "dorothy3"}))
	})

	It("can use Stream prefix to return a Stream", func() {
		st := userRepository.StreamAllByName("Dorothy")
		count := st.Count()
		Expect(count).To(BeNumerically("==", 1))
	})

	It("can return a Stream with Error", func() {
		st, err := userRepository.FindStreamWithErrorByName("Dorothy")
		Expect(err).NotTo(HaveOccurred())
		Expect(st.Count()).To(BeNumerically("==", 1))

		stEmpty, err := userRepository.FindStreamWithErrorByName("NonExistent")
		Expect(err).NotTo(HaveOccurred())
		Expect(stEmpty.Count()).To(BeNumerically("==", 0))
	})

	It("rejects invalid Stream return type with mismatched entity", func() {
		Expect(userRepository.findBadStreamByName).To(BeNil())
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
	transactionManager TransactionManager      `@:"Inject"`
	gormTransactional  GormTransactional       `@:"Inject"`
	userRepository     UserRepositoryInterface `@:"Inject"`
}

func (this *UserService) TransactionalChangeRollback() {
	err := this.transactionManager.Transactional(func() Error {
		user, _ := this.userRepository.FindById(1).Get()
		user.Name = "bar"
		_, err := this.userRepository.Save(user)
		Expect(err).NotTo(HaveOccurred())
		return NewError("TransactionalError", "Rolling back")
	})
	Expect(err).To(HaveOccurred())

	u := this.userRepository.FindByName("bar")
	Expect(u).To(BeNil())
}

func (this *UserService) TransactionalChange() {
	this.transactionManager.SetContext(context.Background())
	defer this.transactionManager.ClearContext()

	err := this.transactionManager.Transactional(func() Error {
		user, _ := this.userRepository.FindById(1).Get()
		user.Name = "bar"
		_, err := this.userRepository.Save(user)
		Expect(err).NotTo(HaveOccurred())
		return nil
	})
	Expect(err).NotTo(HaveOccurred())

	u := this.userRepository.FindByName("bar")
	Expect(u).NotTo(BeNil())
	Expect(u.Name).To(Equal("bar"))
}

func (this *UserService) ErrorAfterRollback() {
	Expect(this.transactionManager.Start()).NotTo(HaveOccurred())
	user, _ := this.userRepository.FindById(1).Get()
	Expect(this.transactionManager.Rollback()).NotTo(HaveOccurred())
	Expect(this.userRepository.FindById(1).IsPresent()).To(BeFalse())
	Expect(this.userRepository.Delete(user)).To(HaveOccurred())
	Expect(this.userRepository.DeleteById(1)).To(HaveOccurred())
	_, err := this.userRepository.Save(user)
	Expect(err).To(HaveOccurred())
	_, err = this.userRepository.SaveAll([]*User{user})
	Expect(err).To(HaveOccurred())
	Expect(this.gormTransactional.Error()).To(HaveOccurred())
	Expect(this.gormTransactional.Commit()).To(HaveOccurred())
	Expect(this.gormTransactional.Rollback()).To(HaveOccurred())
	this.transactionManager.End()
}

func (this *UserService) RepositoryErrorRollsBackTransaction() {
	err := this.transactionManager.Transactional(func() Error {
		user, _ := this.userRepository.FindById(1).Get()
		user.Name = "bar"
		_, saveErr := this.userRepository.Save(user)
		Expect(saveErr).NotTo(HaveOccurred())

		// FindWithErrorByUserName on non-existent user reports error to GormTransactional -> TransactionManager
		_, findErr := this.userRepository.FindWithErrorByUserName("nonexistent")
		Expect(findErr).To(HaveOccurred())
		Expect(this.transactionManager.Error()).To(Equal(findErr))

		// Subsequent Enlist or Commit is rejected; even returning nil from closure must fail and rollback
		return nil
	})
	Expect(err).To(HaveOccurred())

	u := this.userRepository.FindByName("bar")
	Expect(u).To(BeNil())
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

	It("automatically rolls back when a repository query reports an error", func() {
		userService.RepositoryErrorRollsBackTransaction()
	})

	It("handles driver start errors when underlying sql.DB is closed", func() {
		sqlDB, err := userRepository.Query().DB()
		Expect(err).NotTo(HaveOccurred())
		Expect(sqlDB.Close()).NotTo(HaveOccurred())

		txErr := userService.transactionManager.Transactional(func() Error {
			_ = userService.userRepository.FindByName("Dorothy")
			return nil
		})
		Expect(txErr).To(HaveOccurred())
	})
})

type tenantCtxKey string

const tenantKey tenantCtxKey = "tenant"

type TypeSplitDBConfig struct {
	Configuration
	userDB   *gorm.DB
	animalDB *gorm.DB
}

func (this *TypeSplitDBConfig) PostConstruct() Error {
	this.userDB = openTestSQLiteDB()
	this.animalDB = openTestSQLiteDB()
	Expect(this.userDB.AutoMigrate(&User{})).NotTo(HaveOccurred())
	Expect(this.animalDB.AutoMigrate(&Animal{})).NotTo(HaveOccurred())
	return nil
}

func (this *TypeSplitDBConfig) DB(ctx context.Context) *gorm.DB {
	if OrmEntityTypeFromContext(ctx) == reflect.TypeFor[Animal]() {
		return this.animalDB
	}
	return this.userDB
}

type RepoTypeSplitDBConfig struct {
	Configuration
	userDB   *gorm.DB
	animalDB *gorm.DB
}

func (this *RepoTypeSplitDBConfig) PostConstruct() Error {
	this.userDB = openTestSQLiteDB()
	this.animalDB = openTestSQLiteDB()
	Expect(this.userDB.AutoMigrate(&User{})).NotTo(HaveOccurred())
	Expect(this.animalDB.AutoMigrate(&Animal{})).NotTo(HaveOccurred())
	return nil
}

func (this *RepoTypeSplitDBConfig) DB(ctx context.Context) *gorm.DB {
	if OrmRepositoryTypeFromContext(ctx) == reflect.TypeFor[*AnimalRepository]() {
		return this.animalDB
	}
	return this.userDB
}

type ContextSplitUserRepository struct {
	OrmRepository[User, uint]
	foundry       Foundry[GormTransactional] `@:"Inject"`
	findByName    func(string) *User
	returnNilInTx bool
}

var _ GormTransactionalResolver = &ContextSplitUserRepository{}

func (this *ContextSplitUserRepository) GormTransactional(ctx context.Context) GormTransactional {
	if this.returnNilInTx {
		return nil
	}
	tx, err := this.foundry.Forge(ctx)
	Expect(err).NotTo(HaveOccurred())
	return tx
}

func (this *ContextSplitUserRepository) FindByName(name string) *User {
	return this.findByName(name)
}

type ContextSplitDBConfig struct {
	Configuration
	defaultDB *gorm.DB
	tenantADB *gorm.DB
	tenantBDB *gorm.DB
}

func (this *ContextSplitDBConfig) PostConstruct() Error {
	this.defaultDB = openTestSQLiteDB()
	this.tenantADB = openTestSQLiteDB()
	this.tenantBDB = openTestSQLiteDB()
	Expect(this.defaultDB.AutoMigrate(&User{})).NotTo(HaveOccurred())
	Expect(this.tenantADB.AutoMigrate(&User{})).NotTo(HaveOccurred())
	Expect(this.tenantBDB.AutoMigrate(&User{})).NotTo(HaveOccurred())
	return nil
}

func (this *ContextSplitDBConfig) DB(ctx context.Context) *gorm.DB {
	switch ctx.Value(tenantKey) {
	case "tenantA":
		return this.tenantADB
	case "tenantB":
		return this.tenantBDB
	default:
		return this.defaultDB
	}
}

func openTestSQLiteDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{
		Logger: sloggorm.New(
			sloggorm.WithHandler(GetLogger().Handler()),
		),
	})
	Expect(err).NotTo(HaveOccurred())
	return db
}

var _ = Describe("Type Splitting & Context Splitting", func() {
	It("splits *gorm.DB drivers by Entity Type T and coordinates multi-DB commit and rollback", func() {
		GetContainer().Reset()

		userRepo := &UserRepository{}
		animalRepo := &AnimalRepository{}
		AddBeamFor(&TypeSplitDBConfig{})
		AddBeamFor(userRepo)
		AddBeamFor(animalRepo)
		GetContainer().EnableScanning()
		Expect(GetContainer().Initialize()).NotTo(HaveOccurred())

		txMgr := GetBeamDefaultInstance[TransactionManager]()

		// 1. Coordinated multi-DB Commit across userDB and animalDB
		Expect(txMgr.Transactional(func() Error {
			u, err := userRepo.Save(&User{Name: "Dorothy", UserName: "dorothy", Age: 25})
			Expect(err).NotTo(HaveOccurred())
			_, err = animalRepo.Save(&Animal{Name: "Toto", UserID: u.ID})
			Expect(err).NotTo(HaveOccurred())
			return nil
		})).NotTo(HaveOccurred())

		Expect(userRepo.FindById(1).IsPresent()).To(BeTrue())
		Expect(animalRepo.FindById(1).IsPresent()).To(BeTrue())

		// 2. Coordinated multi-DB Rollback: error after saving in both DBs rolls back BOTH userDB and animalDB
		Expect(txMgr.Transactional(func() Error {
			_, err := userRepo.Save(&User{Name: "Glinda", UserName: "glinda", Age: 30})
			Expect(err).NotTo(HaveOccurred())
			_, err = animalRepo.Save(&Animal{Name: "Lion", UserID: 2})
			Expect(err).NotTo(HaveOccurred())
			return NewError("AbortBoth", "rollback both databases")
		})).To(HaveOccurred())

		Expect(userRepo.FindByName("Glinda")).To(BeNil())
		Expect(animalRepo.FindById(2).IsPresent()).To(BeFalse())
	})

	It("splits *gorm.DB drivers by Repository Type", func() {
		GetContainer().Reset()

		userRepo := &UserRepository{}
		animalRepo := &AnimalRepository{}
		AddBeamFor(&RepoTypeSplitDBConfig{})
		AddBeamFor(userRepo)
		AddBeamFor(animalRepo)
		GetContainer().EnableScanning()
		Expect(GetContainer().Initialize()).NotTo(HaveOccurred())

		_, err := userRepo.Save(&User{Name: "Oz", UserName: "oz", Age: 50})
		Expect(err).NotTo(HaveOccurred())
		_, err = animalRepo.Save(&Animal{Name: "Tiger", UserID: 1})
		Expect(err).NotTo(HaveOccurred())

		Expect(userRepo.FindById(1).IsPresent()).To(BeTrue())
		Expect(animalRepo.FindById(1).IsPresent()).To(BeTrue())

		// Also verify helper functions on empty context and direct ParentType/ParentInstance context
		Expect(OrmRepositoryTypeFromContext(context.Background())).To(BeNil())
		Expect(OrmEntityTypeFromContext(context.Background())).To(BeNil())
		parentCtx := context.WithValue(context.Background(), PrototypeContextParentType, reflect.TypeFor[*UserRepository]())
		parentCtx = context.WithValue(parentCtx, PrototypeContextParentInstance, userRepo)
		Expect(OrmRepositoryTypeFromContext(parentCtx)).To(Equal(reflect.TypeFor[*UserRepository]()))
		Expect(OrmEntityTypeFromContext(parentCtx)).To(Equal(reflect.TypeFor[User]()))
	})

	It("splits *gorm.DB drivers dynamically by context.Context via GormTransactionalResolver and Foundry", func() {
		GetContainer().Reset()

		repo := &ContextSplitUserRepository{}
		AddBeamFor(&ContextSplitDBConfig{})
		AddBeamFor(repo)
		GetContainer().EnableScanning()
		Expect(GetContainer().Initialize()).NotTo(HaveOccurred())

		txMgr := GetBeamDefaultInstance[TransactionManager]()

		// Tenant A transaction
		ctxA := context.WithValue(context.Background(), tenantKey, "tenantA")
		txMgr.SetContext(ctxA)
		Expect(txMgr.Transactional(func() Error {
			_, err := repo.Save(&User{Name: "AliceA", UserName: "alice_a", Age: 20})
			Expect(err).NotTo(HaveOccurred())
			// Second call in same context hits goroutine txCache
			Expect(repo.FindByName("AliceA")).NotTo(BeNil())
			return nil
		})).NotTo(HaveOccurred())
		txMgr.ClearContext()

		// Tenant B transaction
		ctxB := context.WithValue(context.Background(), tenantKey, "tenantB")
		txMgr.SetContext(ctxB)
		Expect(txMgr.Transactional(func() Error {
			// AliceA does not exist in Tenant B's database
			Expect(repo.FindByName("AliceA")).To(BeNil())
			_, err := repo.Save(&User{Name: "BobB", UserName: "bob_b", Age: 30})
			return err
		})).NotTo(HaveOccurred())
		txMgr.ClearContext()

		// Verify isolation between Tenant A, Tenant B, and Default DB
		txMgr.SetContext(ctxA)
		Expect(repo.FindByName("AliceA")).NotTo(BeNil())
		Expect(repo.FindByName("BobB")).To(BeNil())
		txMgr.ClearContext()

		txMgr.SetContext(ctxB)
		Expect(repo.FindByName("BobB")).NotTo(BeNil())
		Expect(repo.FindByName("AliceA")).To(BeNil())
		txMgr.ClearContext()

		// Fallback to default transactionManager when resolver returns nil
		repo.returnNilInTx = true
		Expect(repo.FindByName("AliceA")).To(BeNil())
	})
})
