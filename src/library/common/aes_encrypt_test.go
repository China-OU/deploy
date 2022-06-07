package common

import "testing"

/*
 * 单元测试：  go test aes_encrypt_test.go aes_encrypt.go   (go test -v xxxxx)
 * 单元测试覆盖率：  go test --cover aes_encrypt_test.go aes_encrypt.go
 * 并发性能测试： go test -bench=. aes_encrypt_test.go aes_encrypt.go
 */


/*
 * 密码测试
 * CFB加密方式：不需要指定长度，加密结果长度与密码长度直相关
 * CBC加密方式：需要16位，如果密码不够需要指定到16位，密码生成的不一样，安全性高
 */
func TestAesEncry(t *testing.T) {
	str := ""
	strEncrypted := AesEncrypt(str)
	t.Log("Encrypted:",strEncrypted)
	strDecrypted := AesDecrypt("ad2876ab137f928599b84e37dbf243be")
	t.Log("Decrypted:",strDecrypted)
}

func TestAesWebEncry(t *testing.T) {
	str := ""
	strEncrypted := WebPwdEncrypt(str)
	t.Log("Web Encrypted:",strEncrypted)
	strDecrypted := WebPwdDecrypt("")
	t.Log("Web Decrypted:",strDecrypted)
}

// 性能测试
func BenchmarkCFBEncry(b *testing.B) {
	//key := "w2j83aH68HzhQwtG"
	//pwd_suffix := "ETGOYgQh"
	//ets := CBCEncrypter("12345678" + pwd_suffix, key)
	//b.Log(string(ets))
	//b.Log("======================================================")
	//dets := CBCDecrypter("703672f5e4bc747b559309924ccde7760fe543d0c2b553e6f0408498ec8be2fc", key)
	//b.Log(dets[:8])
}

// 并发性能测试
func BenchmarkCFBEncryParallel(b *testing.B) {
	//key := "w2j83aH68HzhQwtG"
	//pwd_suffix := "ETGOYgQh"
	//ets := CBCEncrypter("12345678" + pwd_suffix, key)
	//b.Log(string(ets))
	//b.Log("======================================================")
	//dets := CBCDecrypter("703672f5e4bc747b559309924ccde7760fe543d0c2b553e6f0408498ec8be2fc", key)
	//b.Log(dets[:8])
}
