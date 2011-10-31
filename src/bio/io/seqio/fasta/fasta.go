// Copyright ©2011 Dan Kortschak <dan.kortschak@adelaide.edu.au>
//
//   This program is free software: you can redistribute it and/or modify
//   it under the terms of the GNU General Public License as published by
//   the Free Software Foundation, either version 3 of the License, or
//   (at your option) any later version.
//
//   This program is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//   GNU General Public License for more details.
//
//   You should have received a copy of the GNU General Public License
//   along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
package fasta

import (
	"os"
	"io"
	"bufio"
	"bytes"
	"bio"
	"bio/seq"
	"bio/util"
)

// Fasta sequence format reader type.
type Reader struct {
	f         io.ReadCloser
	r         *bufio.Reader
	IDPrefix  []byte
	SeqPrefix []byte
	last      []byte
}

// Returns a new fasta format reader using f.
func NewReader(f io.ReadCloser) *Reader {
	return &Reader{
		f:         f,
		r:         bufio.NewReader(f),
		IDPrefix:  []byte(">"), // default delimiters
		SeqPrefix: []byte(""),  // default delimiters
		last:      nil,
	}
}

// Returns a new fasta format reader using a filename.
func NewReaderName(name string) (r *Reader, err os.Error) {
	var f *os.File
	if f, err = os.Open(name); err != nil {
		return
	}
	return NewReader(f), nil
}

// Read a single sequence and return it or an error.
func (self *Reader) Read() (sequence *seq.Seq, err os.Error) {
	var line, label, body []byte
	label = self.last

READ:
	for {
		if line, err = self.r.ReadBytes('\n'); err == nil {
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			line = bytes.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			switch {
			case bytes.HasPrefix(line, self.IDPrefix):
				if self.last == nil {
					label = line[len(self.IDPrefix):]
					self.last = label
				} else {
					label = self.last
					self.last = line[len(self.IDPrefix):] // entering a new sequence so exit read loop
					break READ
				}
			case bytes.HasPrefix(line, self.SeqPrefix):
				line = bytes.Join(bytes.Fields(line[len(self.SeqPrefix):]), nil)
				body = append(body, line...)
			}
		} else {
			if self.last != nil {
				self.last = nil
				err = nil
				break
			} else {
				return nil, os.EOF
			}
		}
	}

	sequence = seq.New(label, body, nil)

	return
}

// Rewind the reader.
func (self *Reader) Rewind() (err os.Error) {
	if s, ok := self.f.(io.Seeker); ok {
		self.last = nil
		_, err = s.Seek(0, 0)
	} else {
		err = bio.NewError("Not a Seeker", 0, self)
	}
	return
}

// Close the reader.
func (self *Reader) Close() (err os.Error) {
	return self.f.Close()
}

// Fasta sequence format writer type.
type Writer struct {
	f         io.WriteCloser
	w         *bufio.Writer
	IDPrefix  string
	SeqPrefix string
	Width     int
}

// Returns a new fasta format writer using f.
func NewWriter(f io.WriteCloser, width int) *Writer {
	return &Writer{
		f:         f,
		w:         bufio.NewWriter(f),
		IDPrefix:  ">", // default delimiters
		SeqPrefix: "",  // default delimiters
		Width:     width,
	}
}

// Returns a new fasta format writer using a filename, truncating any existing file.
// If appending is required use NewWriter and os.OpenFile.
func NewWriterName(name string, width int) (w *Writer, err os.Error) {
	var f *os.File
	if f, err = os.Create(name); err != nil {
		return
	}
	return NewWriter(f, width), nil
}

// Write a single sequence and return the number of bytes written and any error.
func (self *Writer) Write(s *seq.Seq) (n int, err os.Error) {
	var ln int
	if n, err = self.w.WriteString(self.IDPrefix + string(s.ID) + "\n"); err == nil {
		for i := 0; i*self.Width <= s.Len(); i++ {
			endLinePos := util.Min(self.Width*(i+1), s.Len())
			ln, err = self.w.WriteString(self.SeqPrefix + string(s.Seq[self.Width*i:endLinePos]) + "\n")
			n += ln
			if err != nil {
				break
			}
		}
	}

	return
}

// Close the writer, flushing any unwritten sequence.
func (self *Writer) Close() (err os.Error) {
	if err = self.w.Flush(); err != nil {
		return
	}
	return self.f.Close()
}