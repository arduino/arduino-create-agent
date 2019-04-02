package tools

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"golang.org/x/crypto/openpgp"

	"github.com/arduino/arduino-create-agent/utilities"
	"github.com/blang/semver"
	"github.com/xrash/smetrics"
)

type system struct {
	Host     string `json:"host"`
	URL      string `json:"url"`
	Name     string `json:"archiveFileName"`
	CheckSum string `json:"checksum"`
}

type tool struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Systems     []system `json:"systems"`
	url         string
	destination string
}

type index struct {
	Packages []struct {
		Name  string `json:"name"`
		Tools []tool `json:"tools"`
	} `json:"packages"`
}

var systems = map[string]string{
	"linuxamd64":   "x86_64-linux-gnu",
	"linux386":     "i686-linux-gnu",
	"darwinamd64":  "apple-darwin",
	"windows386":   "i686-mingw32",
	"windowsamd64": "i686-mingw32",
	"linuxarm":     "arm-linux-gnueabihf",
}

func mimeType(data []byte) (string, error) {
	return http.DetectContentType(data[0:512]), nil
}

// gpg --export YOURKEYID --export-options export-minimal,no-export-attributes | hexdump /dev/stdin -v -e '/1 "%02X"'
var publicKeyHex string = "99020D0452FAA2FA011000D0C5604932111750628F171E4E612D599ABEA8E4309888B9B9E87CCBD3AAD014B27454B0AF08E7CDD019DA72D492B6CF882AD7FA8571E985C538582DA096C371E7FCD95B71BC00C0E92BDDC26801F1B11C86814E0EA849E5973F630FC426E6A5F262C22986CB489B5304005202BA729D519725E3E6042C9199C8ECE734052B7376CF40A864679C3594C93203EBFB3F82CD42CD956F961792233B4C7C1A28252360F48F1D6D8662F2CF93F87DB40A99304F61828AF8A3EB07239E984629DC0B1D5C6494C9AFB5C8F8B9A53F1324C254A1AEA4CD9219AB4DF8667653AC9A6E76C3DB37CE8F60B546F78ECA43A90CB82A2278A291D2E98D66753B56F0595280C6E33B274F631846806D97447DD5C9438E7EC85779D9FA2173E088CE6FA156E291FAFD432C4FC2B1EB251DAFD13C721FF6618F696B77C122FB75E3CBCB446FAAA7FFFDD071C81C6C3D2360D495964C623617352915BBB30FA7C903EA096BF01E77FC84C8B51C02EB11BC9F03F19C7E81254F1F786E64A7A9F7F438873583CFA40F49C0F041432EAECCEC7EE9BA465A30320306F0E2E65EBE01E533CBBD8B1C1C04222213D5D05F4B689193DB60A68A6F1FC8B2ADD9E58A104E482AAD3869CCC42236EDC9CBE8561E105837AB6F7A764DCE5D8CB62608E8133F0FDD5F8FAFBE3BC57EE551ADC7386AADD443B331716EC032ACF9C639BF9DFE62301D4F197221F10DEF0011010001B42041726475696E6F204C4C43203C737570706F72744061726475696E6F2E63633E890238041301020022050252FAA2FA021B03060B090807030206150802090A0B0416020301021E01021780000A09107BAF404C2DFAB4AEF8090FFE20C3B36BF786D692969DA2ECFD7BCA3961E735D3CBB5585D7AB04BB8A0B64B64528ED76DB4752FA24523AA1E07B69A6A66CDDAE074A6A572800228194DD5916A956BF22606D866C7FD81F32878E06FEC200DDB0703D805E1A61006EB0B5BDB3AA89C095BB259BD93C7AAE8BDB18468A6DBE30F85BD6A3271F5456EB22BC2BCE99DB3A054D9BCA8F562C01B99E6BF4C2136B62771EEF54CB2AE95F8E2FE543284C37EB77E5104D49939ABAEF323CA5F1A66CA48ED423DBB3A2CFF12792CCA71ACD1E3032186CC7D05A13E0D66A3258E443527AAF921B7EA70C6CC10E2A51FCAB4DD130A10D3D29B1B01FB4207EF6501D3A9186BDB652ECCC9F354599A114DD3F80F9ED3493AC51A5C4F1F3BB59049EE7EC61411E90E02F27789E87B18A860551DFDFFA870E8542F6128E167CE1875C5C5B1128259347B85265487006B173AA631F1CDA1EDC68C54978E1D0FE3B310CC0F49F9AE84F37B1472437B69DA125BAFDC99AE57C2245F70747E1EFD52849C40469247CF13CB679A31AF4700468E09ED1ECFE5A53F67C80C48A0B0C1334FAE9650584DFD406ADA30FFBEED659256D40924432B029BBB24CEF22195D389381F0B1EB964C6494942335E74A373D869D1FB0C7967F30F79D71AB06929CEBB660514C2567284BD9EC32470B263539B3AFF5D3FBA9A275D4665E6B502B4031B63F511C1DFDD16B617A6FB046FCEB018A7A01CEFB9020D0452FAA2FA011000D6DE1747395EB3836103D30FA5CF555F6FBC982FB8B0FD72389CD6E99A88ACA1BCBD8BAD35211929AB5AB7F656BA1AFFA8C9A5AF83436FC8FE36AB403453E3E6EC679371AD81657FA1506956B1165D8887E3FB7EF366EFCCA82EE543E0B22170D0164A6702EF5280398A901CB6262E63C0AE378FD8CA1957EEED9CE48AA3D481BD117A2CA0341C3E16FE20CB6A5C3130A19B364F656CDC45E2216DE7ACFAD429967D71D101CADE10BA64F4075801ED2E9E3A3293114543456A26236CCA459DC7700D2E9C692BADCA9BA0CDE7189CD594B20CA4D1F20A70B02B9B50F70CFC6F7697B1D500702CE29492C7CD28C5D555475788DDE57482BC39E8465A720E25866AC931D5D7030AB61136BF702B25BC850A5089D1E6F0F68B8AE894ADFC3C92BB836888E3DB5A940426DBE7BBC5BDD3DDD6F5123627D1CE6FD1845CC66A920094391BE783069CB05746C0A55DAFC869FDAF0A08F81099E4F4CD07D05C7269C538C341CF1EDB94114B8CD97B44214EA58EEDB93FAB772013A1D77A08B9208082F9617A6CFE39B56F0078406C6267ABF5CF1078C49B1AB9B60EA1451351CF889EF72D7D696B23B22F753B28979AF10237B579A350FA5596A3B22244FA91402562AE530E814EF19A9E3448F465F78C16220DE0663F7B97C7F0EF1629E2F64A76B21BB695A3DE505B22B09B3459A3CE2180424BD67C8482EBD5EBC8128F98634EEE8707001101000189021F041801020009050252FAA2FA021B0C000A09107BAF404C2DFAB4AE050B1000C1434E8CC0D6F8E60E2FB091AA5EA04E7612B29D3823E09914F704DE1835A7B202D3F619183BD3A16439BFA31A6AF342672E8F59184333C4F56D18AF3B7CE8326F655F7C8DD1D4B38A1964E6A4D7550D159CE1B5EC44BC2091B1097CABE724C0E8C4942C2CF82672E3F209322270D133313CF601E07756B705946A45235DAF7294BCD34292D989EFDFDA2F46AF0AEAEC72F55DC6B2940C7C6A409B7FAD3354D5CA30C3E4EE29F9218A56EF8D7FBA2A7BB8E6304110A21DF0C847C4B761CDE408CE156D53091535A800C1C522CA33C71105B11550A145FD0E41B464146B46D46F08DFAEF9B03D313D54A1E4A82E8749895AB78521DAA8E66EEF6F7B17A0CA4B4CBFCB937713B9806269556EBD88AE87996EFAC0846ACBA0D3412FC0A5E90923C261CD443E4D6C1AE93D83166937C5F606A14FD73DB4919A0ED416D4B3163420F57FACCE9C9347BD5501BE3FC830472B64068E5FF5B09E7425030625246720D21608DEE829F84E8365527F764C91DA93372C72AA4054B458104CAFC2BDCED63DC80F36E7BD4BE0D3A19E20E3FED90F80F9E1584853B971B8E847C27027123B9AA19C3E90B41B3A643D3D5BE2FC134ADA8396D072D37E7101B64CE83E1802D0D5DDA9150B6C21564987950C9601FC2147F139C7A9906640A0883981B452F25AF7A0F32FAA2148ECDD9B04B93AFCED00F11AA0E6695C2F92676B8DB9E93172FD7779B9020D04530B05A7011000CAA1A8FF4BF9D0F0AC9EDBCA3B4D26E3E569DFEA04341F3E6ACE23AE5D87D62C2600DFF10B106144A1B52FF8B695A590D65C681F69DEE454800C235160EBE3FC1436193E1278D56C86E2BBB2187BEAAC1E1D04D40A392B1457471D10A2B6BF39CDF35D1A090A9406BCB06BDEF83A12A490C5E17D68884AD2686042669E2B458AD3CC0377DDA9C58D7070CE29A53E0E7C876D61B29A2DE2A9D73F914D0FF3B0E35E2ED361B60A8C3C3D4C7E77E17A939283BFDA2EC5725A2BFAAC18C6A64ACBEC776760D7086EA42BD93031E8B59FB8DFEFF77E5F80DBEB84ADE74B3A6F9E4D0F3140A8D0F576ED00548883C85271AA7F2450D1061F56CB839786038861D5A2473B7F58EBC00D2BB9EFEB1A2DF612A7B9087C326FBB08F2879102253316784272967A886089D61D5AB0FDB33737D35F27C2886ABB4D4E88F541D0BBAD04AEF7BD3ED66A1282B762BD6F8EEDC3760773B157C1A2D4E4586E43B28879C54E7599F9A34E1524E6E7F9B8EA13CC5A2DF5C1920AF74833EDDEC8EB9A8BE33196702DFD656D81ACBBFE3A10DA882EAA3065D9C9476C0A7B66C15D0063CB7AD1A2EB31537CB443F21B81642436943FE6C45E6AF9C2B595D4DFCB64B83F2CA6B4DD536726C6EC4761A340C18E32B2D7210640B9AB1D8E2165C0DD38BC9FD9DB6A30B380DF08C3F10002A6636FDC79CD2312B606F5F116AC665618A56BBE46C494FC7E23C7001101000189043E0418010200090502530B05A7021B02022909107BAF404C2DFAB4AEC15D200419010200060502530B05A7000A091024A26BAD7F29429187700FFE30ED1B7C96B3846AC7B363F9602D2886F7913A9C451C31E043AD75597024D460B59E6A60A6EE3D58E656901237A2465F8402169A816B38170AF550284EB420B7E827386D66852D68125A27FA6770F139EE7FCAEF43000673B7C7D168614877603C875AD593E333AE9237DB77065FB8375CE98FA1BF7FB1733034AAC61F1D23A3EFF8665702C10968C7991458F88D151B3448C7D9334059431A63D30A9C8E636A99D88DA8DB04CB8C64F1183AC873FF0942EF9555B6B3F192AD5F221AC9737F875CCAE21E88EC45CB35E40C0FF1AAF0A8FE44876D93A930A03CC4846A29102C956F39F2AC5808CCBCD7F4868A8E8E8B9A66EA18C275CEF9C371AB0592796ED57D757A3BAB31FF8E3887F6041E61BDA433E7D68CB2D5F28E81F57843D5032D73BF67119C137FC4CE8BEF4F705D690E47A530B1A85B8B6A09A4AE16A2973C11D69031B89BE92B0751DB7FE74F6F1C219C8B93E5C68EC1403856DF28E96E27737A7FB9C80F6EE9EC485A0609DC4EB8DF444F61C76A97F32ADFA2D8B4784DF3ABA4DE1B57894B9CF89934A143451308D73CF79ECC8BF382B8A34F24DC335238D8353767B363F5432D9A81C84F7D2FAB6E36E7188FA911120A905C67342A996251EBECAC13BD543A9B3C2C063AE294FDD15C66D5DD9224F3E936325F525700F2129D0B31CE8CCD4EBA5DEDB89F0A2BFC0C43E732F695161E4F33CE5DED14B1E98654547B110FFF7CBC2BA513721A96DD18964635069343FA8EEF4D492BFA55C930F9C78DF1F7454F1BDD40F4B04BDE9F9B9A9923A303D96D0CBFA361921AFEF13AED098D0CF70E84C0DDB20C58821351D2359B131671AAF5D2484717A4CAF385DB0CC19FBC37A3FC04F4F387D6934C1E84B9C1291231A14F69A1BF6708875C7DE00E3EFE3C7855A2459C96245C5F0D21FC00E87A0C18F80A3B79C0E28EA27493309C535254421BE7CDFBEFB5B44DAEA56B6859430FCCBEE766048F891AD5CB503866B98E521ED69B37E4165012A45E29836E2A0380728C1108E4C8A32EA186E1A855F78DA5506B6CF86DB888A87FAB6E15A90E3416469522DF5BD8872D729B35E6D82C974CD80076C26008015AB216C83FAF64E488F07D2BD01F51B0963F87BE0AB8392B442227BF7215148038B0C55189024D7C1B032DB1B3C56C66953E530C5B323634FC584A476CAD285EF1108011D14D9D180A75A9DFC936AFC7EF9E6C3F3CFEDD894894CE60358E7156B3A65ED7644DEA343A133F5D4DE4D33B74281086A0C20515AC4151CFED93C56DD574E578FDEE72C4115C25CAEC5EAD97C147F27F4EAE67FEFFEA0DC1CDF5D636AC331CB74DF477C9C3B3706F9DAF50C2E13AC8DE8CC9DD3C79E59EC779EE489D915CF22FDC53E3B3C7710FE8368DF11B9ACDF5F3CAE1F43CB7312E5E9F57F248692B3681CBA3E49207878FD33ED2A47CE9CE9B4E4A6EFD8F0AD2CD"

func checkGPGSig(fileName string, sigFileName string) error {

	// Get a Reader for the signature file
	sigFile, err := os.Open(sigFileName)
	if err != nil {
		return err
	}
	defer sigFile.Close()

	// Get a Reader for the signature file
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	publicKeyBin, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return err
	}

	keyring, _ := openpgp.ReadKeyRing(bytes.NewReader(publicKeyBin))

	_, err = openpgp.CheckDetachedSignature(keyring, file, sigFile)

	return err
}

func (t *Tools) DownloadPackageIndex(index_file, signature_file string) error {
	// Fetch the index
	resp, err := http.Get(t.IndexURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read the body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Fetch the signature
	signature, err := http.Get(t.IndexURL + ".sig")
	if err != nil {
		return err
	}
	defer signature.Body.Close()

	// Read the body
	signature_body, err := ioutil.ReadAll(signature.Body)
	if err != nil {
		return err
	}
	ioutil.WriteFile(index_file, body, 0644)
	ioutil.WriteFile(signature_file, signature_body, 0644)

	t.LastRefresh = time.Now()

	return nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// Download will parse the index at the indexURL for the tool to download.
// It will extract it in a folder in .arduino-create, and it will update the
// Installed map.
//
// pack contains the packager of the tool
// name contains the name of the tool.
// version contains the version of the tool.
// behaviour contains the strategy to use when there is already a tool installed
//
// If version is "latest" it will always download the latest version (regardless
// of the value of behaviour)
//
// If version is not "latest" and behaviour is "replace", it will download the
// version again. If instead behaviour is "keep" it will not download the version
// if it already exists.
func (t *Tools) Download(pack, name, version, behaviour string) error {

	index_file := path.Join(t.Directory, "package_index.json")
	signature_file := path.Join(t.Directory, "package_index.json.sig")

	if _, err := os.Stat(path.Join(t.Directory, "package_index.json")); err != nil || time.Since(t.LastRefresh) > 1*time.Hour {
		// Download the file again and save it
		err = t.DownloadPackageIndex(index_file, signature_file)
		if err != nil {
			return err
		}
	}

	err := checkGPGSig(index_file, signature_file)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadFile(index_file)
	if err != nil {
		return err
	}

	var data index
	json.Unmarshal(body, &data)

	// Find the tool by name
	correctTool, correctSystem := findTool(pack, name, version, data)

	if correctTool.Name == "" || correctSystem.URL == "" {
		t.Logger("We couldn't find a tool with the name " + name + " and version " + version + " packaged by " + pack)
		return nil
	}

	key := correctTool.Name + "-" + correctTool.Version

	// Check if it already exists
	if behaviour == "keep" {
		location, ok := t.installed[key]
		if ok && pathExists(location) {
			// overwrite the default tool with this one
			t.installed[correctTool.Name] = location
			t.Logger("The tool is already present on the system")
			return t.writeMap()
		}
	}

	// Download the tool
	t.Logger("Downloading tool " + name + " from " + correctSystem.URL)
	resp, err := http.Get(correctSystem.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read the body
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Checksum
	checksum := sha256.Sum256(body)
	checkSumString := "SHA-256:" + hex.EncodeToString(checksum[:sha256.Size])

	if checkSumString != correctSystem.CheckSum {
		return errors.New("Checksum doesn't match")
	}

	// Decompress
	t.Logger("Unpacking tool " + name)

	location := path.Join(dir(), pack, correctTool.Name, correctTool.Version)
	err = os.RemoveAll(location)

	if err != nil {
		return err
	}

	srcType, err := mimeType(body)
	if err != nil {
		return err
	}

	switch srcType {
	case "application/zip":
		location, err = extractZip(t.Logger, body, location)
	case "application/x-bz2":
	case "application/octet-stream":
		location, err = extractBz2(t.Logger, body, location)
	case "application/x-gzip":
		location, err = extractTarGz(t.Logger, body, location)
	default:
		return errors.New("Unknown extension for file " + correctSystem.URL)
	}

	if err != nil {
		t.Logger("Error extracting the archive: " + err.Error())
		return err
	}

	err = t.installDrivers(location)
	if err != nil {
		return err
	}

	// Ensure that the files are executable
	t.Logger("Ensure that the files are executable")

	// Update the tool map
	t.Logger("Updating map with location " + location)

	t.installed[name] = location
	t.installed[name+"-"+correctTool.Version] = location
	return t.writeMap()
}

func findTool(pack, name, version string, data index) (tool, system) {
	var correctTool tool
	correctTool.Version = "0.0"

	for _, p := range data.Packages {
		if p.Name != pack {
			continue
		}
		for _, t := range p.Tools {
			if version != "latest" {
				if t.Name == name && t.Version == version {
					correctTool = t
				}
			} else {
				// Find latest
				v1, _ := semver.Make(t.Version)
				v2, _ := semver.Make(correctTool.Version)
				if t.Name == name && v1.Compare(v2) > 0 {
					correctTool = t
				}
			}
		}
	}

	// Find the url based on system
	var correctSystem system
	max_similarity := 0.7

	for _, s := range correctTool.Systems {
		similarity := smetrics.Jaro(s.Host, systems[runtime.GOOS+runtime.GOARCH])
		if similarity > max_similarity {
			correctSystem = s
			max_similarity = similarity
		}
	}

	return correctTool, correctSystem
}

func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

func commonPrefix(sep byte, paths []string) string {
	// Handle special cases.
	switch len(paths) {
	case 0:
		return ""
	case 1:
		return path.Clean(paths[0])
	}

	c := []byte(path.Clean(paths[0]))

	// We add a trailing sep to handle: common prefix directory is included in the path list
	// (e.g. /home/user1, /home/user1/foo, /home/user1/bar).
	// path.Clean will have cleaned off trailing / separators with
	// the exception of the root directory, "/" making it "//"
	// but this will get fixed up to "/" below).
	c = append(c, sep)

	// Ignore the first path since it's already in c
	for _, v := range paths[1:] {
		// Clean up each path before testing it
		v = path.Clean(v) + string(sep)

		// Find the first non-common byte and truncate c
		if len(v) < len(c) {
			c = c[:len(v)]
		}
		for i := 0; i < len(c); i++ {
			if v[i] != c[i] {
				c = c[:i]
				break
			}
		}
	}

	// Remove trailing non-separator characters and the final separator
	for i := len(c) - 1; i >= 0; i-- {
		if c[i] == sep {
			c = c[:i]
			break
		}
	}

	return string(c)
}

func removeStringFromSlice(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func findBaseDir(dirList []string) string {
	if len(dirList) == 1 {
		return path.Dir(dirList[0]) + "/"
	}

	// https://github.com/backdrop-ops/contrib/issues/55#issuecomment-73814500
	dontdiff := []string{"pax_global_header"}
	for _, v := range dontdiff {
		dirList = removeStringFromSlice(dirList, v)
	}

	commonBaseDir := commonPrefix('/', dirList)
	if commonBaseDir != "" {
		commonBaseDir = commonBaseDir + "/"
	}
	return commonBaseDir
}

func extractZip(log func(msg string), body []byte, location string) (string, error) {
	path, err := utilities.SaveFileonTempDir("tooldownloaded.zip", bytes.NewReader(body))
	r, err := zip.OpenReader(path)
	if err != nil {
		return location, err
	}

	var dirList []string

	for _, f := range r.File {
		dirList = append(dirList, f.Name)
	}

	basedir := findBaseDir(dirList)
	log(fmt.Sprintf("selected baseDir %s from Zip Archive Content: %v", basedir, dirList))

	for _, f := range r.File {
		fullname := filepath.Join(location, strings.Replace(f.Name, basedir, "", -1))
		log(fmt.Sprintf("generated fullname %s removing %s from %s", fullname, basedir, f.Name))
		if f.FileInfo().IsDir() {
			os.MkdirAll(fullname, f.FileInfo().Mode().Perm())
		} else {
			os.MkdirAll(filepath.Dir(fullname), 0755)
			perms := f.FileInfo().Mode().Perm()
			out, err := os.OpenFile(fullname, os.O_CREATE|os.O_RDWR, perms)
			if err != nil {
				return location, err
			}
			rc, err := f.Open()
			if err != nil {
				return location, err
			}
			_, err = io.CopyN(out, rc, f.FileInfo().Size())
			if err != nil {
				return location, err
			}
			rc.Close()
			out.Close()

			mtime := f.FileInfo().ModTime()
			err = os.Chtimes(fullname, mtime, mtime)
			if err != nil {
				return location, err
			}
		}
	}
	return location, nil
}

func extractTarGz(log func(msg string), body []byte, location string) (string, error) {
	bodyCopy := make([]byte, len(body))
	copy(bodyCopy, body)
	tarFile, _ := gzip.NewReader(bytes.NewReader(body))
	tarReader := tar.NewReader(tarFile)

	var dirList []string

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		dirList = append(dirList, header.Name)
	}

	basedir := findBaseDir(dirList)
	log(fmt.Sprintf("selected baseDir %s from TarGz Archive Content: %v", basedir, dirList))

	tarFile, _ = gzip.NewReader(bytes.NewReader(bodyCopy))
	tarReader = tar.NewReader(tarFile)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			//return location, err
		}

		path := filepath.Join(location, strings.Replace(header.Name, basedir, "", -1))
		info := header.FileInfo()

		// Create parent folder
		dirmode := info.Mode() | os.ModeDir | 0700
		if err = os.MkdirAll(filepath.Dir(path), dirmode); err != nil {
			return location, err
		}

		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return location, err
			}
			continue
		}

		if header.Typeflag == tar.TypeSymlink {
			err = os.Symlink(header.Linkname, path)
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			continue
		}
		_, err = io.Copy(file, tarReader)
		if err != nil {
			//return location, err
		}
		file.Close()
	}
	return location, nil
}

func extractBz2(log func(msg string), body []byte, location string) (string, error) {
	bodyCopy := make([]byte, len(body))
	copy(bodyCopy, body)
	tarFile := bzip2.NewReader(bytes.NewReader(body))
	tarReader := tar.NewReader(tarFile)

	var dirList []string

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		dirList = append(dirList, header.Name)
	}

	basedir := findBaseDir(dirList)
	log(fmt.Sprintf("selected baseDir %s from Bz2 Archive Content: %v", basedir, dirList))

	tarFile = bzip2.NewReader(bytes.NewReader(bodyCopy))
	tarReader = tar.NewReader(tarFile)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			continue
			//return location, err
		}

		path := filepath.Join(location, strings.Replace(header.Name, basedir, "", -1))
		info := header.FileInfo()

		// Create parent folder
		dirmode := info.Mode() | os.ModeDir | 0700
		if err = os.MkdirAll(filepath.Dir(path), dirmode); err != nil {
			return location, err
		}

		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return location, err
			}
			continue
		}

		if header.Typeflag == tar.TypeSymlink {
			err = os.Symlink(header.Linkname, path)
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			continue
			//return location, err
		}
		_, err = io.Copy(file, tarReader)
		if err != nil {
			//return location, err
		}
		file.Close()
	}
	return location, nil
}

func (t *Tools) installDrivers(location string) error {
	OK_PRESSED := 6
	extension := ".bat"
	preamble := ""
	if runtime.GOOS != "windows" {
		extension = ".sh"
		// add ./ to force locality
		preamble = "./"
	}
	if _, err := os.Stat(filepath.Join(location, "post_install"+extension)); err == nil {
		t.Logger("Installing drivers")
		ok := MessageBox("Installing drivers", "We are about to install some drivers needed to use Arduino/Genuino boards\nDo you want to continue?")
		if ok == OK_PRESSED {
			os.Chdir(location)
			t.Logger(preamble + "post_install" + extension)
			oscmd := exec.Command(preamble + "post_install" + extension)
			if runtime.GOOS != "linux" {
				// spawning a shell could be the only way to let the user type his password
				TellCommandNotToSpawnShell(oscmd)
			}
			err = oscmd.Run()
			return err
		} else {
			return errors.New("Could not install drivers")
		}
	}
	return nil
}

func makeExecutable(location string) error {
	location = path.Join(location, "bin")
	files, err := ioutil.ReadDir(location)
	if err != nil {
		return err
	}

	for _, file := range files {
		err = os.Chmod(path.Join(location, file.Name()), 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
