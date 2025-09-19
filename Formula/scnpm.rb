class Scnpm < Formula
  desc "Security scanner for malware-affected npm packages"
  homepage "https://github.com/GigacoreLLC/scnpm"
  url "https://github.com/GigacoreLLC/scnpm/archive/refs/tags/v1.0.0.tar.gz"
  sha256 "b8af19cf21efdf416ef6ddfe116ac791efc789b53820611f63a17bfd1835fa2b"
  license "MIT"
  head "https://github.com/GigacoreLLC/scnpm.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w -X main.version=#{version}"), "./main.go"
  end

  test do
    # Test help command
    assert_match "Security scanner for malware-affected npm packages", shell_output("#{bin}/scnpm --help")
    
    # Test that it mentions badpak.json in help
    assert_match "badpak.json", shell_output("#{bin}/scnpm --help")
  end
end
