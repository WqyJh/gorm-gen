package tests_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"gorm.io/gen/tests/diy_method"
)

const (
	generateDirPrefix = ".gen/"
	expectDirPrefix   = ".expect/"
)

var _ = os.Setenv("GORM_DIALECT", "mysql")

type User struct {
	Id       string    `gorm:"primaryKey"`
	Posts    []Post    `gorm:"foreignKey:AuthorId"`
	Comments []Comment `gorm:"foreignKey:AuthorId"`
}

type Post struct {
	Id       string `gorm:"primaryKey"`
	AuthorId string
	Author   User      `gorm:"foreignKey:Id"`
	Comments []Comment `gorm:"foreignKey:PostId"`
}

type Comment struct {
	Id       string `gorm:"primaryKey"`
	PostId   string
	Post     Post `gorm:"foreignKey:Id"`
	AuthorId string
	Author   User `gorm:"foreignKey:Id"`
}

var generateCase = map[string]func(dir string) *gen.Generator{
	generateDirPrefix + "dal_1": func(dir string) *gen.Generator {
		g := gen.NewGenerator(gen.Config{
			OutPath: dir + "/query",
			Mode:    gen.WithDefaultQuery,
		})
		g.UseDB(DB)
		g.ApplyBasic(g.GenerateAllTable()...)
		return g
	},
	generateDirPrefix + "dal_2": func(dir string) *gen.Generator {
		g := gen.NewGenerator(gen.Config{
			OutPath: dir + "/query",
			Mode:    gen.WithDefaultQuery,

			WithUnitTest: true,

			FieldNullable:     true,
			FieldCoverable:    true,
			FieldWithIndexTag: true,
		})
		g.UseDB(DB)
		g.WithJSONTagNameStrategy(func(c string) string { return "-" })
		g.ApplyBasic(g.GenerateAllTable()...)
		return g
	},
	generateDirPrefix + "dal_3": func(dir string) *gen.Generator {
		g := gen.NewGenerator(gen.Config{
			OutPath: dir + "/query",
			Mode:    gen.WithDefaultQuery | gen.WithQueryInterface,

			WithUnitTest: true,

			FieldNullable:     true,
			FieldCoverable:    true,
			FieldWithIndexTag: true,
		})
		g.UseDB(DB)
		g.WithJSONTagNameStrategy(func(c string) string { return "-" })
		g.ApplyBasic(g.GenerateAllTable(gen.FieldGORMTagReg(".", func(tag field.GormTag) field.GormTag {
			//tag.Set("serialize","json")
			tag.Remove("comment")
			return tag
		}))...)
		return g
	},
	generateDirPrefix + "dal_4": func(dir string) *gen.Generator {
		g := gen.NewGenerator(gen.Config{
			OutPath: dir + "/query",
			Mode:    gen.WithDefaultQuery | gen.WithQueryInterface,

			WithUnitTest: true,

			FieldNullable:     true,
			FieldCoverable:    true,
			FieldWithIndexTag: true,
		})
		g.UseDB(DB)
		g.WithJSONTagNameStrategy(func(c string) string { return "-" })
		g.ApplyBasic(g.GenerateAllTable()...)
		g.ApplyInterface(func(testIF diy_method.TestIF, testFor diy_method.TestFor, method diy_method.InsertMethod, selectMethod diy_method.SelectMethod) {
		}, g.GenerateModel("users"))
		return g
	},
	generateDirPrefix + "dal_5": func(dir string) *gen.Generator {
		g := gen.NewGenerator(gen.Config{
			OutPath: dir + "/query",
			Mode:    gen.WithDefaultQuery | gen.WithQueryInterface,

			WithUnitTest: true,

			FieldNullable:     true,
			FieldCoverable:    true,
			FieldWithIndexTag: true,
		})
		g.UseDB(DB)
		g.WithJSONTagNameStrategy(func(c string) string { return "-" })
		g.ApplyBasic(g.GenerateModel("users", gen.WithMethod(diy_method.TestForWithMethod{})))
		return g
	},
	generateDirPrefix + "dal_6": func(dir string) *gen.Generator {
		g := gen.NewGenerator(gen.Config{
			OutPath: dir + "/query",
			Mode:    gen.WithDefaultQuery | gen.WithQueryInterface,

			WithUnitTest: true,

			FieldNullable:     true,
			FieldCoverable:    true,
			FieldWithIndexTag: true,
		})
		g.UseDB(DB)
		g.WithJSONTagNameStrategy(func(c string) string { return "-" })
		g.ApplyBasic(g.GenerateModelAs("users", DB.Config.NamingStrategy.SchemaName("users"), gen.WithMethod(diy_method.TestForWithMethod{})))
		return g
	},
	generateDirPrefix + "dal_7": func(dir string) *gen.Generator {
		g := gen.NewGenerator(gen.Config{
			OutPath: dir + "/query",
			Mode:    gen.WithDefaultQuery | gen.WithQueryInterface,

			WithUnitTest: true,

			FieldNullable:     true,
			FieldCoverable:    true,
			FieldWithIndexTag: true,
		})
		g.UseDB(DB)
		g.WithJSONTagNameStrategy(func(c string) string { return "-" })

		banks := g.GenerateModel("banks")
		creditCards := g.GenerateModel("credit_cards")
		customers := g.GenerateModel("customers",
			gen.FieldRelate(field.HasOne, "Bank", banks, &field.RelateConfig{
				JSONTag: "bank",
				GORMTag: field.GormTag{
					"foreignKey": []string{"BankID"},
					"references": []string{"ID"},
				},
			}),
			gen.FieldRelate(field.HasMany, "CreditCards", creditCards, &field.RelateConfig{
				JSONTag: "credit_cards",
				GORMTag: field.GormTag{
					"foreignKey": []string{"CustomerRefer"},
					"references": []string{"ID"},
				},
			}),
		)
		g.ApplyBasic(customers)
		return g
	},
	generateDirPrefix + "dal_8": func(dir string) *gen.Generator {
		g := gen.NewGenerator(gen.Config{
			OutPath: dir + "/query",
			Mode:    gen.WithDefaultQuery | gen.WithQueryInterface,

			WithUnitTest: true,

			FieldNullable:     true,
			FieldCoverable:    true,
			FieldWithIndexTag: true,
		})
		g.ApplyBasic(User{}, Post{}, Comment{})

		return g
	},
}

func TestGenerate(t *testing.T) {
	for dir := range generateCase {
		t.Run("TestGenerate_"+dir, func(dir string) func(t *testing.T) {
			return func(t *testing.T) {
				t.Parallel()
				if err := matchGeneratedFile(dir); err != nil {
					t.Errorf("generated file is unexpected: %s", err)
				}
			}
		}(dir))
	}
}

func matchGeneratedFile(dir string) error {
	_ = os.Remove(dir + "/query/gen_test.db")

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	expectDir := expectDirPrefix + strings.TrimPrefix(dir, generateDirPrefix)
	diffResult, err := exec.CommandContext(ctx, "diff", "-r", expectDir, dir).CombinedOutput()
	if err != nil {
		return fmt.Errorf("diff %s and %s got: %w\n%s", expectDir, dir, err, diffResult)
	}
	if len(diffResult) != 0 {
		return fmt.Errorf("unexpected content: %s", diffResult)
	}
	return nil
}

func TestGenerate_expect(t *testing.T) {
	if os.Getenv("GEN_EXPECT") == "" {
		t.SkipNow()
	}
	g := gen.NewGenerator(gen.Config{
		OutPath: expectDirPrefix + "dal_test" + "/query",
		Mode:    gen.WithDefaultQuery,
	})
	g.UseDB(DB)
	g.ApplyBasic(g.GenerateAllTable()...)
	g.Execute()

	g = gen.NewGenerator(gen.Config{
		OutPath: expectDirPrefix + "dal_test_relation" + "/query",
		Mode:    gen.WithDefaultQuery,
	})
	g.UseDB(DB)

	banks := g.GenerateModel("banks")
	creditCards := g.GenerateModel("credit_cards")
	customers := g.GenerateModel("customers",
		gen.FieldRelate(field.HasOne, "Bank", banks, &field.RelateConfig{
			JSONTag: "bank",
			GORMTag: field.GormTag{
				"foreignKey": []string{"BankID"},
				"references": []string{"ID"},
			},
		}),
		gen.FieldRelate(field.HasMany, "CreditCards", creditCards, &field.RelateConfig{
			JSONTag: "credit_cards",
			GORMTag: field.GormTag{
				"foreignKey": []string{"CustomerRefer"},
				"references": []string{"ID"},
			},
		}),
	)
	g.ApplyBasic(customers, creditCards, banks)
	g.Execute()
}

func Test_GenSkipImpl(t *testing.T) {
	dir := ".gen/skip_impl_test"
	os.RemoveAll(dir)
	g := gen.NewGenerator(gen.Config{
		OutPath: dir + "/query",
		Mode:    gen.WithDefaultQuery | gen.WithQueryInterface,
	})
	g.UseDB(DB)
	model := g.GenerateModel("users")
	g.ApplyInterface(func(diy_method.TestSkipImpl) {}, model)
	g.Execute()

	queryFile := dir + "/query/users.gen.go"
	content, err := os.ReadFile(queryFile)
	if err != nil {
		t.Fatalf("read generated file failed: %v", err)
	}
	str := string(content)
	if !strings.Contains(str, "func (u userDo) NoSkipMethod(") {
		t.Error("should generate NoSkipMethod implementation")
	}
	if strings.Contains(str, "func (u userDo) SkipMethod(") || strings.Contains(str, "func (u usersDo) SkipMethod(") {
		t.Error("should not generate SkipMethod implementation for // gen:skip interface")
	}
}
