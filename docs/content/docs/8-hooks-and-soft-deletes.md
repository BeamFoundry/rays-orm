---
title: "8. Hooks & Soft Deletes"
weight: 8
---

# Lifecycle Hooks & Soft Deletes

Understanding how GORM interacts with the Rays IoC container is crucial for structuring your
enterprise business logic securely.

## 1. Model Hooks vs. Service Logic (The IoC Boundary)

GORM provides native model hooks (e.g., `BeforeCreate`, `AfterSave`). It is tempting to place all your
pre-save logic in these hooks, but you must respect the **IoC Boundary**.

Models (like `User{}`) are instantiated by the database driver, not by the Rays IoC container.
Therefore, **you cannot use the `@:"Inject"` tag inside a GORM model.**

### The Rule of Thumb

* **Static/Pure Logic** (UUID generation, default timestamps) belongs in a **Model Hook**.
* **Dynamic/Impure Logic** (Password Hashing via an injected `PasswordEncoder`, Audit Logging) belongs
in the **Service Layer**.

**Good Practice: Static Logic in the Model**
```go
func (u *User) BeforeCreate(tx *gorm.DB) error {
    // Perfectly fine: No injected dependencies required!
    if u.UUID == "" {
        u.UUID = uuid.New().String()
    }
    return nil
}
```

**Good Practice: Dependency Logic in the Service**
```go
type UserService struct {
    Service
    repo    *UserRepository  `@:"Inject"`
    encoder *PasswordEncoder `@:"Inject"` // We need this!
}

func (s *UserService) Register(email, rawPassword string) Error {
    // The hashing requires an injected dependency, so it MUST happen
    // in the Service layer before calling the repository.
    hashed, _ := s.encoder.Hash(rawPassword)

    user := &User{Email: email, Password: hashed}
    _, err := s.repo.Save(user)

    return err
}
```

## 2. Soft Deletes

Enterprise applications rarely execute physical `DELETE` statements. Maintaining a historical audit
trail requires "Soft Deletes".

Because Rays ORM wraps GORM natively, soft deletes are fully supported without any extra configuration.
If your model struct embeds `gorm.Model` or explicitly includes a `gorm.DeletedAt` field, the framework adapts automatically.

```go
import "gorm.io/gorm"

type User struct {
    gorm.Model // Automatically includes ID, CreatedAt, UpdatedAt, and DeletedAt
    Email string
}
```

When you call the standard base repository methods:
* `repo.Delete(user)`
* `repo.DeleteById(userId)`

Rays ORM will automatically execute an `UPDATE` statement setting the `deleted_at` column, rather than a
hard `DELETE` statement. Subsequent auto-generated queries (`FindByName`) will automatically filter out soft-deleted records.
