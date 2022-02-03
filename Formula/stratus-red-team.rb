# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class StratusRedTeam < Formula
  desc ""
  homepage "https://stratus-red-team.cloud"
  version "1.3.0"
  license "Apache-2.0"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/DataDog/stratus-red-team/releases/download/v1.3.0/stratus-red-team_1.3.0_Darwin_x86_64.tar.gz"
      sha256 "5cc4a5f0d417cf02ee18781d40c9ed4556aa4bdc985ec0638a5f5fdbc9a3d27b"

      def install
        bin.install "stratus"
      end
    end
    if Hardware::CPU.arm?
      url "https://github.com/DataDog/stratus-red-team/releases/download/v1.3.0/stratus-red-team_1.3.0_Darwin_arm64.tar.gz"
      sha256 "9c6ecc47cd096acdd0286178c7f0a2aeddf028b899d6b1380a6825e0d2e18aee"

      def install
        bin.install "stratus"
      end
    end
  end

  on_linux do
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/DataDog/stratus-red-team/releases/download/v1.3.0/stratus-red-team_1.3.0_Linux_arm64.tar.gz"
      sha256 "3763850453ba364ce985530f88e629ae11f335e5504a1d4aceec2038852739a1"

      def install
        bin.install "stratus"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/DataDog/stratus-red-team/releases/download/v1.3.0/stratus-red-team_1.3.0_Linux_x86_64.tar.gz"
      sha256 "c40414480e1ac3ad0e7192bf1626173b5c2531d76beb473f1dfbf7b5dceb22bf"

      def install
        bin.install "stratus"
      end
    end
  end
end
