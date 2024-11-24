function isSelectionType(t) {
  return ['checkbox', 'radio'].includes(t)
}

function isTextType(t) {
  return ['text', 'paragraph'].includes(t)
}

function scrollIntoViewIfNeeded(el) {
  clientRect = el.getBoundingClientRect()
  if (clientRect.bottom > window.innerHeight) {
    el.scrollIntoView(false);
  }
  
  if (clientRect.top < 0) {
    el.scrollIntoView();
  }
}

document.addEventListener('alpine:init', () => {
  Alpine.data('form', () => ({
    title: '',
    description: '',
    taskName: 'Task Name',
    taskSummary: 'Task Summary',
    fields: [],
    counter: 0,

    init() {
    },

    newField(f, idx) {
      this.fields.splice(idx, 0, f)
      this.counter++;
    },

    newCheckbox() {
      return {
        id: this.counter,
        type: 'checkbox',
        prompt: 'Prompt ' + (this.counter + 1),
        required: false,
        options: ['Option 1', 'Option 2'],
        selectedOptions: [],
      };
    },

    newRadio() {
      return {
        id: this.counter,
        type: 'radio',
        prompt: 'Prompt ' + (this.counter + 1),
        required: false,
        options: ['Option 1', 'Option 2'],
        selectedOption: [],
      }
    },

    newText() {
      return {
        id: this.counter,
        type: 'text',
        prompt: 'Prompt ' + (this.counter + 1),
        required: false,
        content: '',
      }
    },

    newParagraph() {
      return {
        id: this.counter,
        type: 'paragraph',
        prompt: 'Prompt ' + (this.counter + 1),
        required: false,
        content: '',
      }
    },

    async submit(username, service) {
      try {
        resp = await api.post(`/app/${username}/services/${service}/form`, {
          "task_name": this.taskName,
          "task_summary": this.taskSummary,
          "fields": this.fields.map((f) => {
            res = {
              type: f.type,
            }
            switch(f.type) {
              case 'checkbox':
                res.filled = f.selectedOptions.length != 0
                res.data = {
                  selectedOptions: f.selectedOptions
                }
                break
              case 'radio':
                res.filled = f.selectedOption != ""
                res.data = {
                  selectedOption: f.selectedOption
                }
                break
              case 'text':
                res.filled = f.content != ""
                res.data = {
                    content: f.content
                }
                break
              case 'paragraph':
                res.type = 'text'
                res.filled = f.content != ""
                res.data = {
                    content: f.content
                }
                break
              default:
                throw new Error(`Invalid type ${f.type}`)
            }
            return res
          })
        });

        console.log(resp)
        if (resp.redirected) {
          window.location.replace(resp.url);
        }
      } catch(e) {
        throw new Error(e)
      }
    },

    async loadFromURL(username, service) {
      try {
        resp = await api.get(`/api/users/${username}/services/${service}`);
        json = await resp.json();

        this.title = json.name;
        this.description = json.description;
        this.fields = []

        for (const f of json.fields) {
          var newField;
          switch (f.type) {
            case 'checkbox':
              newField = this.newCheckbox();
              newField.options = f.data.options
              break;
            case 'radio':
              newField = this.newRadio();
              newField.options = f.data.options
              break;
            case 'text':
              if (f.data.paragraph) {
                newField = this.newParagraph();
              } else {
                newField = this.newText();
              }
              break;
            default:
              throw new Error(`Invalid type ${f.type}`);
          }

          newField.prompt = f.prompt;
          newField.required = f.required;

          this.fields.push(newField);
          this.counter++;
        }

        await this.$nextTick();
      } catch(e) {
        throw new Error(e)
      }
    },
  }))

})
