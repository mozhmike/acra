// Copyright 2016, Cossack Labs Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package zone

import (
	"container/list"
)

type KeyChecker interface {
	HasZonePrivateKey([]byte) bool
}

type ZoneIdMatcher struct {
	matched     bool
	matchers    *list.List
	zoneId      []byte
	matcherPool *MatcherPool
	keychecker  KeyChecker
}

func NewZoneMatcher(matcherPool *MatcherPool, keychecker KeyChecker) *ZoneIdMatcher {
	matcher := &ZoneIdMatcher{
		matchers:    list.New(),
		matcherPool: matcherPool,
		matched:     false,
		keychecker:  keychecker,
	}
	matcher.addEmptyMatcher()
	return matcher
}

func (zoneMatcher *ZoneIdMatcher) IsMatched() bool {
	return zoneMatcher.matched
}

func (zoneMatcher *ZoneIdMatcher) Reset() {
	zoneMatcher.matched = false
	zoneMatcher.clearMatchers()
}

func (zoneMatcher *ZoneIdMatcher) GetZoneId() []byte {
	if zoneMatcher.IsMatched() {
		return zoneMatcher.zoneId
	}
	return []byte{}
}

func (zoneMatcher *ZoneIdMatcher) SetMatched(id []byte) {
	zoneMatcher.zoneId = id
	zoneMatcher.matched = true
}

func (zoneMatcher *ZoneIdMatcher) Match(c byte) bool {
	currentElement := zoneMatcher.matchers.Front()
	var toRemove *list.Element
	var matcher Matcher
	isMatched := false
	for {
		matcher = currentElement.Value.(Matcher)
		if matcher.Match(c) {
			if matcher.IsMatched() {
				if zoneMatcher.keychecker.HasZonePrivateKey(matcher.GetZoneId()) {
					zoneMatcher.zoneId = matcher.GetZoneId()
					zoneMatcher.matched = true
					isMatched = true
				}
				toRemove = currentElement
			} else {
				isMatched = true
			}
		} else {
			// if no match and it's not last matcher, delete him
			if currentElement != zoneMatcher.matchers.Back() {
				toRemove = currentElement
			}
		}
		// if last matcher (previously was empty) has match, add empty matcher and quit
		if currentElement == zoneMatcher.matchers.Back() && matcher.HasAnyMatch() {
			zoneMatcher.addEmptyMatcher()
			if toRemove != nil {
				zoneMatcher.remove(toRemove)
			}
			break
		}

		currentElement = currentElement.Next()
		if toRemove != nil {
			zoneMatcher.remove(toRemove)
			toRemove = nil
		}
		if currentElement == nil {
			break
		}
	}
	return isMatched
}

func (zoneMatcher *ZoneIdMatcher) remove(element *list.Element) {
	zoneMatcher.matchers.Remove(element)
	zoneMatcher.matcherPool.Release(element.Value.(Matcher))
}

func (zoneMatcher *ZoneIdMatcher) clearMatchers() {
	/* delete all matcher except the last that should be emptyMatcher */
	var previous *list.Element
	element := zoneMatcher.matchers.Front()
	for {
		if element.Next() != nil {
			previous = element
			element = element.Next()
			zoneMatcher.remove(previous)
		} else {
			return
		}
	}
}

func (zoneMatcher *ZoneIdMatcher) addEmptyMatcher() {
	matcher := zoneMatcher.matcherPool.Acquire()
	zoneMatcher.matchers.PushBack(matcher)
}
