"""Tests for agent_army.editor."""

from __future__ import annotations

import pytest

from agent_army.editor import select_multi, select_one


class TestSelectOne:
    """select_one() — single-choice numbered menu."""

    def test_valid_choice(self, monkeypatch: pytest.MonkeyPatch) -> None:
        monkeypatch.setattr("builtins.input", lambda _: "2")
        result = select_one("Pick:", ["alpha", "beta", "gamma"])
        assert result == "beta"

    def test_first_choice(self, monkeypatch: pytest.MonkeyPatch) -> None:
        monkeypatch.setattr("builtins.input", lambda _: "1")
        result = select_one("Pick:", ["only"])
        assert result == "only"

    def test_invalid_then_valid(self, monkeypatch: pytest.MonkeyPatch) -> None:
        responses = iter(["0", "abc", "2"])
        monkeypatch.setattr("builtins.input", lambda _: next(responses))
        result = select_one("Pick:", ["a", "b"])
        assert result == "b"


class TestSelectMulti:
    """select_multi() — multi-choice numbered menu."""

    def test_single_selection(self, monkeypatch: pytest.MonkeyPatch) -> None:
        monkeypatch.setattr("builtins.input", lambda _: "1")
        result = select_multi("Pick:", ["a", "b", "c"])
        assert result == ["a"]

    def test_multiple_selection(self, monkeypatch: pytest.MonkeyPatch) -> None:
        monkeypatch.setattr("builtins.input", lambda _: "1,3")
        result = select_multi("Pick:", ["a", "b", "c"])
        assert result == ["a", "c"]

    def test_invalid_then_valid(self, monkeypatch: pytest.MonkeyPatch) -> None:
        responses = iter(["99", "1,2"])
        monkeypatch.setattr("builtins.input", lambda _: next(responses))
        result = select_multi("Pick:", ["x", "y"])
        assert result == ["x", "y"]
