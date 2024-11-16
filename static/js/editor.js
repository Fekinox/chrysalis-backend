function isSelectionType(t) {
  return ['checkbox', 'radio'].includes(t)
}

function isTextType(t) {
  return ['text', 'paragraph'].includes(t)
}

function newOption(v) {
  return {
    value: v,
    currentValue: v,
  };
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
  Alpine.data('editor', () => ({
    title: 'Untitled form',
    description: '',
    fields: [],
    counter: 0,

    init() {
      Sortable.create(this.$refs.fieldList, {
        animation: 150,
        handle: '.field-handle',
        onEnd: (ev) => {
          const [movedItem] = this.fields.splice(ev.oldIndex, 1);
          this.fields.splice(ev.newIndex, 0, movedItem);

          // need to manually reset the _x_prevKeys field used by alpine's
          // x-for directive to prevent it from going out of sync
          // https://github.com/alpinejs/alpine/discussions/4157
          this.$refs.fieldList.querySelector('template')._x_prevKeys =
            this.fields.map((x) => x.id);
        }
      });

      this.newField(this.newCheckbox(), 0)
    },

    remove_field(i) {
      this.fields = this.fields.filter((f) => f.id != i);
    },

    async newField(f, idx) {
      this.fields.splice(idx, 0, f)
      this.counter++;
      await this.$nextTick()
      scrollIntoViewIfNeeded(this.$refs.fieldList.querySelectorAll(':scope > div').item(idx));
    },

    setField(idx, type) {
      newField = null
      switch(type) {
        case 'checkbox':
          newField = this.newCheckbox()
          break;
        case 'radio':
          newField = this.newRadio()
          break;
        case 'text':
          newField = this.newText()
          break;
        case 'paragraph':
          newField = this.newParagraph()
          break;
        default:
          throw new Exception('invalid type')
      }
      [oldField] = this.fields.splice(idx, 1, newField)

      this.fields[idx].prompt = oldField.prompt
      this.fields[idx].id = oldField.id
      if (isSelectionType(oldField.type) && isSelectionType(newField.type)) {
        this.fields[idx].options = oldField.options
      }
    },

    newCheckbox() {
      return {
        id: this.counter,
        type: 'checkbox',
        prompt: 'Prompt ' + (this.counter + 1),
        required: false,
        options: [newOption('Option 1'), newOption('Option 2')],
      };
    },

    newRadio() {
      return {
        id: this.counter,
        type: 'radio',
        prompt: 'Prompt ' + (this.counter + 1),
        required: false,
        options: [newOption('Option 1'), newOption('Option 2')],
      }
    },

    newText() {
      return {
        id: this.counter,
        type: 'text',
        prompt: 'Prompt ' + (this.counter + 1),
        required: false,
      }
    },

    newParagraph() {
      return {
        id: this.counter,
        type: 'paragraph',
        prompt: 'Prompt ' + (this.counter + 1),
        required: false,
      }
    },

    addOption(fieldIdx, v) {
      this.fields[fieldIdx].options.push({
        value: v,
        currentValue: v,
      })
    },

    async newCheckboxOption(fieldIdx, tgt) {
      numOpts = this.fields[fieldIdx].options.length
      this.addOption(fieldIdx, `Option ${numOpts + 1}`)
      await this.$nextTick();
      fieldContainer = tgt;
      newField =
        fieldContainer.querySelectorAll(':scope > div').item(this.fields[fieldIdx].options.length - 1);
      newField.querySelector(`input[type='text']`).select();
    },

    async newRadioOption(fieldIdx, tgt) {
      numOpts = this.fields[fieldIdx].options.length
      this.addOption(fieldIdx, `Option ${numOpts + 1}`)
      await this.$nextTick();
      fieldContainer = tgt
      newField =
        fieldContainer.querySelectorAll(':scope > div').item(this.fields[fieldIdx].options.length - 1);
      newField.querySelector(`input[type='text']`).focus();
    },

    setOption(fieldIdx, optIdx) {
      value = this.fields[fieldIdx].options[optIdx].value
      cleanedValue = this.fields[fieldIdx].options[optIdx].currentValue.trim()
      if (cleanedValue == '') {
        this.fields[fieldIdx].options[optIdx].currentValue = value
        return
      }
      this.fields[fieldIdx].options[optIdx].value = cleanedValue
    },

    submit() {
      console.log(this.fields.map((x) => x))
    },
  }))
})