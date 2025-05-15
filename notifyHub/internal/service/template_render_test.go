package service

import "testing"

func TestRenderTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		vars map[string]string
		want string
	}{
		{
			name: "replaces known placeholders",
			in:   "Hi {{name}}, welcome to {{product}}!",
			vars: map[string]string{"name": "Ameer", "product": "NotifyHub"},
			want: "Hi Ameer, welcome to NotifyHub!",
		},
		{
			name: "allows whitespace inside braces",
			in:   "Hello {{ name }}",
			vars: map[string]string{"name": "Ameer"},
			want: "Hello Ameer",
		},
		{
			name: "leaves unknown placeholders",
			in:   "Hi {{name}}, code {{code}}",
			vars: map[string]string{"name": "Ameer"},
			want: "Hi Ameer, code {{code}}",
		},
		{
			name: "empty vars returns input",
			in:   "Hi {{name}}",
			vars: nil,
			want: "Hi {{name}}",
		},
		{
			name: "supports underscores in keys",
			in:   "Order {{order_id}}",
			vars: map[string]string{"order_id": "42"},
			want: "Order 42",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := RenderTemplate(tt.in, tt.vars)
			if got != tt.want {
				t.Fatalf("RenderTemplate() = %q, want %q", got, tt.want)
			}
		})
	}
}
